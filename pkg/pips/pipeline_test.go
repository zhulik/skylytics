package pips_test

import (
	"context"
	"log"
	"skylytics/pkg/pips"
	"skylytics/pkg/pips/apply"
	"testing"
)

var (
	subPipe = apply.Pipeline(pips.New[string, string](duplicateMap))

	lenMap = apply.Map(func(ctx context.Context, s string) (int, error) {
		return len(s), nil
	})

	gt6Filter = apply.Filter(func(ctx context.Context, i int) (bool, error) {
		return i > 6, nil
	})
)

func TestPipeline(t *testing.T) {
	t.Parallel()

	res := pips.New[string, int]().
		Then(subPipe).
		Then(lenMap).
		Then(apply.Batch(3)).
		Then(apply.Flatten()).
		Then(gt6Filter).
		Run(t.Context(), inputChan())

	for m := range res {
		log.Println(m.Unpack())
	}
}
