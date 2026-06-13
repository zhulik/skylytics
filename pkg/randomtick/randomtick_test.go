package randomtick

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPauseWithinBounds(t *testing.T) {
	const minPause = 10 * time.Millisecond
	const maxPause = 20 * time.Millisecond

	for range 100 {
		got := Pause(minPause, maxPause)
		if got < minPause || got > maxPause {
			t.Fatalf("Pause() = %v, want within [%v, %v]", got, minPause, maxPause)
		}
	}
}

func TestPauseReturnsMinWhenMaxLessThanMin(t *testing.T) {
	const minPause = 5 * time.Millisecond
	got := Pause(minPause, minPause-time.Millisecond)
	if got != minPause {
		t.Fatalf("Pause() = %v, want %v", got, minPause)
	}
}

func TestTickerTicks(t *testing.T) {
	ticker := New(t.Context(), time.Millisecond, 5*time.Millisecond)

	select {
	case <-ticker.C:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("ticker did not tick")
	}
}

func TestTickerStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	ticker := New(ctx, time.Millisecond, 2*time.Millisecond)

	select {
	case <-ticker.C:
	case <-time.After(20 * time.Millisecond):
		t.Fatal("ticker did not tick before cancel")
	}

	cancel()

	_, ok := <-ticker.C
	if ok {
		t.Fatal("ticker ticked after context cancel")
	}
}

func TestLoopRunsUntilContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var calls atomic.Int32
	done := make(chan struct{})

	go func() {
		Loop(ctx, time.Millisecond, 2*time.Millisecond, func(context.Context) {
			if calls.Add(1) >= 2 {
				cancel()
			}
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("loop did not finish after context cancellation")
	}

	if calls.Load() < 2 {
		t.Fatalf("fn called %d times, want at least 2", calls.Load())
	}
}
