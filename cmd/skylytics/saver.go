package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"skylytics/db/enums"
	dbmodels "skylytics/db/models"
	"skylytics/internal/core"
	"time"

	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/samber/lo"
	"github.com/samber/ro"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/types"
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
				Did:    &event.Did,
				TimeUs: lo.ToPtr(time.Unix(0, event.TimeUS)),
				Kind:   lo.ToPtr(enums.EventKind(event.Kind)),
			}

			if event.Commit != nil {
				setter.Operation = &sql.Null[enums.CommitOperation]{Valid: true, V: enums.CommitOperation(event.Commit.Operation)}
				setter.Collection = &sql.Null[string]{Valid: true, V: event.Commit.Collection}
				setter.Rkey = &sql.Null[string]{Valid: true, V: event.Commit.RKey}
				setter.Record = &sql.Null[types.JSON[json.RawMessage]]{Valid: true, V: types.JSON[json.RawMessage]{Val: event.Commit.Record}}
				setter.Cid = &sql.Null[string]{Valid: true, V: event.Commit.CID}
			}

			if event.Account != nil {
				setter.AccountActive = &sql.Null[bool]{Valid: true, V: event.Account.Active}
				setter.AccountDid = &sql.Null[string]{Valid: true, V: event.Account.Did}
				setter.AccountSeq = &sql.Null[int64]{Valid: true, V: event.Account.Seq}
				if event.Account.Status != nil {
					setter.AccountStatus = &sql.Null[string]{Valid: true, V: *event.Account.Status}
				}

				accountTime, err := time.Parse(time.RFC3339, event.Account.Time)
				if err != nil {
					s.Logger.Error("failed to parse account time", "error", err)
				}

				setter.AccountTime = &sql.Null[time.Time]{Valid: true, V: accountTime}
			}

			if event.Identity != nil {
				setter.IdentityDid = &sql.Null[string]{Valid: true, V: event.Identity.Did}
				if event.Identity.Handle != nil {
					setter.IdentityHandle = &sql.Null[string]{Valid: true, V: *event.Identity.Handle}
				}
				setter.IdentitySeq = &sql.Null[int64]{Valid: true, V: event.Identity.Seq}

				identityTime, err := time.Parse(time.RFC3339, event.Identity.Time)
				if err != nil {
					s.Logger.Error("failed to parse identity time", "error", err)
				}

				setter.IdentityTime = &sql.Null[time.Time]{Valid: true, V: identityTime}
			}

			return setter
		})

		if len(mods) == 0 {
			return
		}

		mods = append(mods, im.OnConflict().DoNothing())

		_, err = dbmodels.Events.Insert(mods...).Exec(ctx, s.DB)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				s.Logger.Error("failed to insert events", "error", err)
			}
		}
	}))
	defer sub.Unsubscribe()

	sub.Wait()

	return nil
}
