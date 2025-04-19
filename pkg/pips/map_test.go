package pips_test

import (
	"context"
	"skylytics/pkg/pips/apply"
	"testing"
)

var (
	duplicateMap = apply.Map(func(ctx context.Context, s string) (string, error) {
		return s + s, nil
	})

	erroredMap = apply.Map(func(ctx context.Context, s string) (string, error) {
		return "", testErr
	})
)

func TestPipeline_Map(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		out := testPipeline(t, duplicateMap)

		requireSuccessfulPiping(t, out, []string{"testtest", "foofoo", "bazzbazz", "traintrain"})
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		out := testPipeline(t, erroredMap)

		requireErroredPiping(t, out, testErr)
	})
}
