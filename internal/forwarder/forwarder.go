package forwarder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

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
				event, record := p.Unpack()

				msg, err := message(event, record)
				if err != nil {
					f.Logger.Error("failed to parse event", "event", event)
					return nil
				}

				_, err = f.JS.PublishMsg(ctx, msg)
				if err != nil {
					f.Logger.Error("failed to publish the event", "event", event)
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

func message(event *core.BlueskyEvent, record *gabs.Container) (*nats.Msg, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	subject := subjectName(event, record)
	msg := nats.NewMsg(subject)
	msg.Data = payload
	msg.Header.Set("did", event.Did)
	msg.Header.Set(nats.MsgIdHdr, fmt.Sprintf("%s-%d", subject, event.TimeUS))

	return msg, nil
}

// Format: event.<base64(did)>.<event_kind>.<suffix>
// Suffix:
// - EventKindAccount: const "active" or "inactive"
// - EventKindIdentity: const "identity"
// - EventKindCommit: "commit.<cid>.<parent_cid>.<root_cid>.<operation>.<collection>"
// 	 if cid, parent_cid or root_cid is missing or is not applicable, "no-cid", "no-parent-cid" and "no-root-cid"
//   are used respectively.

// Example:
// event.ZGlkOnBsYzp0endoZmxrNTI3ZW41b3V3eW9rdnd6ZnI=.commit.bafyreigo3ep3skdmshjtd25snglccjvkqcjjeupp363q2l5npneb5zk2ka.bafyreifvouk5b3ctrj2wue5aufne4s3pobjsf45bmbt2tpaju26jgl4szq.bafyreifvouk5b3ctrj2wue5aufne4s3pobjsf45bmbt2tpaju26jgl4szq.create.app.bsky.feed.post
func subjectName(event *core.BlueskyEvent, record *gabs.Container) string {
	did64 := base64.StdEncoding.EncodeToString([]byte(event.Did))

	suffix := ""

	switch event.Kind {
	case models.EventKindCommit:
		suffix = commitSubjectSuffix(event.Commit, record)
	case models.EventKindAccount:
		if event.Account.Active {
			suffix = "active"
		} else {
			suffix = "inactive"
		}

	case models.EventKindIdentity:
		suffix = "identity"
	}

	return fmt.Sprintf("event.%s.%s.%s", did64, event.Kind, suffix)
}

func commitSubjectSuffix(commit *models.Commit, record *gabs.Container) string {
	cid := commit.CID
	if cid == "" {
		cid = "no-cid"
	}

	parentCID := "no-parent-cid"
	rootCID := "no-root-cid"

	if commit.Collection == "app.bsky.feed.post" && record != nil {
		parent, ok := record.Path("reply.parent.cid").Data().(string)
		if ok {
			parentCID = parent
		}
		root, ok := record.Path("reply.root.cid").Data().(string)
		if ok {
			rootCID = root
		}
	}

	return fmt.Sprintf("%s.%s.%s.%s.%s", cid, parentCID, rootCID, commit.Operation, commit.Collection)
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
