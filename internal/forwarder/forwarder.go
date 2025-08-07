package forwarder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/nats-io/nats.go"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"

	"skylytics/internal/core"

	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	htRE = regexp.MustCompile(`#\w+`)

	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_events_processed_total",
		Help: "The total number of processed events",
	}, []string{"kind", "operation", "status"})

	commitProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_commit_processed_total",
		Help: "The total number of processed commits",
	}, []string{"commit_type", "operation"})

	hashtagProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_commit_hashtags_total",
		Help: "The total number of processed hashtags",
	}, []string{"tag"})
)

type Forwarder struct {
	Logger *slog.Logger
	Sub    core.BlueskySubscriber
	JS     core.JetstreamClient
}

func (f *Forwarder) Run(ctx context.Context) error {
	return f.Sub.ConsumeToPipeline(ctx, pips.New[*core.BlueskyEvent, *core.BlueskyEvent]().
		Then(
			apply.Zip(func(_ context.Context, event *core.BlueskyEvent) (*gabs.Container, error) {
				if event.Commit != nil && len(event.Commit.Record) > 0 {
					return gabs.ParseJSON(event.Commit.Record)
				}
				return nil, nil
			}),
		).
		Then(apply.Each(countEvent)).
		Then(
			apply.Each(func(ctx context.Context, p pips.P[*core.BlueskyEvent, *gabs.Container]) error {
				event, _ := p.Unpack()

				msg, err := message(event)
				if err != nil {
					f.Logger.Error("failed to parse event", "event", event, "error", err)
					return nil
				}

				ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
				defer cancel()

				_, err = f.JS.PublishMsg(ctx, msg)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return nil
					}
					f.Logger.Error("failed to publish the event", "event", event, "error", err)
					return nil
				}
				return nil
			}),
		).
		Then(
			apply.Map(func(_ context.Context, p pips.P[*core.BlueskyEvent, *gabs.Container]) (*core.BlueskyEvent, error) {
				return p.A(), nil
			}),
		),
	)
}

func message(event *core.BlueskyEvent) (*nats.Msg, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	subject := subjectName(event)
	msg := nats.NewMsg(subject)
	msg.Data = payload
	msg.Header.Set(nats.MsgIdHdr, fmt.Sprintf("%s-%d", subject, event.TimeUS))

	return msg, nil
}

// Format: event.<event_kind>.<suffix>
// Suffix:
// - EventKindAccount: const "active" or "inactive"
// - EventKindIdentity: const "identity"
// - EventKindCommit: "<operation>.<collection>.<cid>". If the commit is a reply, the cid is the root cid.
// 	 If the commit is a like of report - cid is the subject cid. if cid or is not applicable, "no-cid" is used.

// Example:
// event.commit.create.app.bsky.feed.post.commit
func subjectName(event *core.BlueskyEvent) string {
	suffix := ""

	switch event.Kind {
	case models.EventKindCommit:
		suffix = fmt.Sprintf("%s.%s", event.Commit.Operation, event.Commit.Collection)
	case models.EventKindAccount:
		if event.Account.Active {
			suffix = "active"
		} else {
			suffix = "inactive"
		}

	case models.EventKindIdentity:
		suffix = "identity"
	}

	return fmt.Sprintf("event.%s.%s", event.Kind, suffix)
}

func countEvent(_ context.Context, p pips.P[*core.BlueskyEvent, *gabs.Container]) error {
	operation := ""
	status := ""

	event, record := p.Unpack()

	switch event.Kind {
	case models.EventKindCommit:
		operation = event.Commit.Operation
		commitProcessed.WithLabelValues(event.Commit.Collection, event.Commit.Operation).Inc()

		if record != nil && event.Commit.Collection == "app.bsky.feed.post" {
			text, ok := record.Path("text").Data().(string)
			if ok {
				tags := hashtags(text)

				if len(tags) > 0 {
					for _, tag := range tags {
						hashtagProcessed.WithLabelValues(tag).Inc()
					}
				}
			}
		}
	case models.EventKindAccount:
		if event.Account.Status != nil {
			status = *event.Account.Status
		}
	case models.EventKindIdentity:
	}

	eventsProcessed.WithLabelValues(event.Kind, operation, status).Inc()

	return nil
}

func hashtags(text string) []string {
	return htRE.FindAllString(text, -1)
}
