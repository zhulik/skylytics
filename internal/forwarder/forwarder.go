package forwarder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Jeffail/gabs"
	"github.com/nats-io/nats.go"
	"github.com/samber/lo"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"

	"skylytics/internal/core"

	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_events_processed_total",
		Help: "The total number of processed events",
	}, []string{"kind", "operation", "status"})

	commitProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_commit_processed_total",
		Help: "The total number of processed commits",
	}, []string{"commit_type", "operation"})
)

type Forwarder struct {
	Logger *slog.Logger
	Sub    core.BlueskySubscriber
	JS     core.JetstreamClient
}

func (f *Forwarder) Run(ctx context.Context) error {
	return f.Sub.ConsumeToPipeline(ctx, pips.New[*core.BlueskyEvent, any]().
		Then(apply.Each(countEvent)).
		Then(
			apply.Each(func(ctx context.Context, event *core.BlueskyEvent) error {
				msg, err := message(event)
				if err != nil {
					f.Logger.Error("failed to parse event", "event", event)
					return nil
				}

				_, err = f.JS.PublishMsg(ctx, msg)
				if err != nil {
					f.Logger.Error("failed to publish the event", "event", event)
					return nil
				}
				return err
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
func subjectName(event *core.BlueskyEvent) string {
	did64 := base64.StdEncoding.EncodeToString([]byte(event.Did))

	suffix := ""

	switch event.Kind {
	case models.EventKindCommit:
		suffix = commitSubjectSuffix(event.Commit)
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

func commitSubjectSuffix(commit *models.Commit) string {
	cid := commit.CID
	if cid == "" {
		cid = "no-cid"
	}

	parentCID := "no-parent-cid"
	rootCID := "no-root-cid"

	if commit.Collection == "app.bsky.feed.post" && len(commit.Record) > 0 {
		parsedRecord := lo.Must(gabs.ParseJSON(commit.Record))

		parent, ok := parsedRecord.Path("reply.parent.cid").Data().(string)
		if ok {
			parentCID = parent
		}
		root, ok := parsedRecord.Path("reply.root.cid").Data().(string)
		if ok {
			rootCID = root
		}
	}

	return fmt.Sprintf("%s.%s.%s.%s.%s", cid, parentCID, rootCID, commit.Operation, commit.Collection)
}

func countEvent(_ context.Context, event *core.BlueskyEvent) error {
	operation := ""
	status := ""

	switch event.Kind {
	case models.EventKindCommit:
		operation = event.Commit.Operation
		commitProcessed.WithLabelValues(event.Commit.Collection, event.Commit.Operation).Inc()
	case models.EventKindAccount:
		if event.Account.Status != nil {
			status = *event.Account.Status
		}
	case models.EventKindIdentity:
	}

	eventsProcessed.WithLabelValues(event.Kind, operation, status).Inc()

	return nil
}
