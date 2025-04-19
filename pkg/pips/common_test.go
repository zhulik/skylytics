package pips_test

import (
	"errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"testing"

	"skylytics/pkg/pips"
)

var (
	testErr = errors.New("test error")
)

func inputChan() <-chan string {
	ch := make(chan string)
	go func() {
		ch <- "test"
		ch <- "foo"
		ch <- "bazz"
		ch <- "train"
		close(ch)
	}()
	return ch
}

func testPipeline(t *testing.T, stages ...pips.Stage) <-chan pips.D[string] {
	t.Helper()

	return pips.New[string, string](stages...).Run(t.Context(), inputChan())
}

func requireSuccessfulPiping[T any](t *testing.T, out <-chan pips.D[T], expected []T) {
	collected := lo.ChannelToSlice(out)

	require.Equal(t, expected,
		lo.Map(collected, func(item pips.D[T], _ int) T {
			require.NoError(t, item.Err)
			return item.Value
		}),
	)
}

func requireErroredPiping[T any](t *testing.T, out <-chan pips.D[T], err error) {
	collected := lo.ChannelToSlice(out)

	require.Len(t, collected, 1)
	require.ErrorIs(t, collected[0].Err, testErr)
}
