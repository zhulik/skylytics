package pips_test

import (
	"context"
	"skylytics/pkg/pips/apply"
	"testing"
)

var (
	lenGt3Filter = apply.Filter(func(ctx context.Context, s string) (bool, error) {
		return len(s) > 3, nil
	})

	erroredFilter = apply.Filter(func(ctx context.Context, s string) (bool, error) {
		return true, testErr
	})
)

func TestPipeline_Filter(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		out := testPipeline(t, lenGt3Filter)

		requireSuccessfulPiping(t, out, []string{"test", "bazz", "train"})
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		out := testPipeline(t, erroredFilter)

		requireErroredPiping(t, out, testErr)
	})
}
