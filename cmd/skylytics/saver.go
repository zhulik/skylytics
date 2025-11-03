package main

import (
	"context"
	"log/slog"
	"skylytics/db/enums"
	dbmodels "skylytics/db/models"
	"skylytics/internal/core"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/samber/lo"
	"github.com/samber/ro"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/im"
)

type saver struct {
	Logger *slog.Logger

	DB core.DB

	Subscriber core.BlueskySubscriber
}

func (s *saver) Run(ctx context.Context) error {
	ch, err := s.Subscriber.Consume(ctx)
	if err != nil {
		return err
	}

	obs := ro.Pipe[*models.Event, []*models.Event](
		ro.FromChannel(ch),
		ro.BufferWithTimeOrCount[*models.Event](100, 1*time.Second),
	)

	sub := obs.Subscribe(ro.OnNext(func(events []*models.Event) {
		mods := lo.Map(events, func(event *models.Event, _ int) bob.Mod[*dialect.InsertQuery] {
			setter := &dbmodels.EventSetter{
				Did:    omit.From(event.Did),
				TimeUs: omit.From(time.Unix(0, event.TimeUS)),
				Kind:   omit.From(enums.EventKind(event.Kind)),
			}

			if event.Commit != nil {
				setter.Operation = omitnull.From(enums.CommitOperation(event.Commit.Operation))
				setter.Collection = omitnull.From(event.Commit.Collection)
				setter.Rkey = omitnull.From(event.Commit.RKey)
				// setter.Record = omitnull.From(event.Commit.Record)
				setter.Cid = omitnull.From(event.Commit.CID)
			}

			if event.Account != nil {
				setter.AccountActive = omitnull.From(event.Account.Active)
				setter.AccountDid = omitnull.From(event.Account.Did)
				setter.AccountSeq = omitnull.From(event.Account.Seq)
				setter.AccountStatus = omitnull.FromPtr(event.Account.Status)

				time, err := time.Parse(time.RFC3339, event.Account.Time)
				if err != nil {
					s.Logger.Error("failed to parse account time", "error", err)
				}

				setter.AccountTime = omitnull.From(time)
			}

			if event.Identity != nil {
				setter.IdentityDid = omitnull.From(event.Identity.Did)
				setter.IdentityHandle = omitnull.FromPtr(event.Identity.Handle)
				setter.IdentitySeq = omitnull.From(event.Identity.Seq)

				time, err := time.Parse(time.RFC3339, event.Identity.Time)
				if err != nil {
					s.Logger.Error("failed to parse identity time", "error", err)
				}

				setter.IdentityTime = omitnull.From(time)
			}

			return setter
		})

		mods = append(mods, im.OnConflict().DoNothing())

		_, err = dbmodels.Events.Insert(mods...).ExecQuery.Exec(ctx, s.DB)
		if err != nil {
			s.Logger.Error("failed to insert events", "error", err)
		}
	}))
	defer sub.Unsubscribe()

	sub.Wait()

	return nil
}
