package pips_test

import (
	"context"
	"log"
	"skylytics/pkg/pips"
	"testing"
)

func TestPipeline(t *testing.T) {
	t.Parallel()

	ch := make(chan string, 10)
	go func() {
		ch <- "test"
		ch <- "foo"
		ch <- "bazz"
		ch <- "train"
		close(ch)
	}()

	subPipe := pips.New[string, string]().
		Then(
			pips.Map(func(ctx context.Context, s string) (string, error) {
				return s + s, nil
			}),
		)

	res := pips.New[string, int]().
		Then(subPipe.ToStage()).
		Then(
			pips.Map(func(ctx context.Context, s string) (int, error) {
				return len(s), nil
			}),
		).
		Then(pips.Batch(3)).
		Then(pips.Flatten()).
		Then(
			pips.Filter(func(ctx context.Context, i int) (bool, error) {
				return i > 6, nil
			}),
		).
		Run(t.Context(), pips.ChanInput(t.Context(), ch))

	for m := range res {
		log.Println(m.Unpack())
	}
}
