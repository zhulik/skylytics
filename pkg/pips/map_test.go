package pips_test

import (
	"context"
	"errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"skylytics/pkg/pips"
	"testing"
)

func TestPipeline_Map(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ch := make(chan string)
		go func() {
			ch <- "test"
			ch <- "foo"
			ch <- "bazz"
			close(ch)
		}()

		out := pips.New[string, string]().
			Then(
				pips.Map(func(ctx context.Context, s string) (string, error) {
					return s + s, nil
				}),
			).Run(t.Context(), pips.ChanInput(t.Context(), ch))

		mapped := lo.ChannelToSlice(out)

		require.Equal(t, []string{"testtest", "foofoo", "bazzbazz"},
			lo.Map(mapped, func(item pips.D[string], _ int) string {
				require.NoError(t, item.Err)
				return item.Value
			}),
		)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		testErr := errors.New("test error")

		ch := make(chan string)
		go func() {
			ch <- "test"
			ch <- "foo"
			ch <- "bazz"
			close(ch)
		}()

		out := pips.New[string, string]().
			Then(
				pips.Map(func(ctx context.Context, s string) (string, error) {
					return "", testErr
				}),
			).Run(t.Context(), pips.ChanInput(t.Context(), ch))

		mapped := lo.ChannelToSlice(out)

		require.Len(t, mapped, 1)
		require.ErrorIs(t, mapped[0].Err, testErr)
	})
}
