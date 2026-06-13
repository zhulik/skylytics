package randomtick

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
)

// Pause returns a uniform random duration in [minPause, maxPause].
func Pause(minPause, maxPause time.Duration) time.Duration {
	if maxPause < minPause {
		return minPause
	}

	span := big.NewInt(int64(maxPause - minPause + 1))
	n, err := rand.Int(rand.Reader, span)
	if err != nil {
		return minPause
	}

	return minPause + time.Duration(n.Int64())
}

// Ticker delivers ticks at randomized intervals in [minPause, maxPause],
// similar to time.Ticker but with a new random pause before each tick.
// It stops when ctx is cancelled.
type Ticker struct {
	C <-chan time.Time
}

// New creates a Ticker that waits a random pause before each tick.
// The ticker stops when ctx is cancelled.
func New(ctx context.Context, minPause, maxPause time.Duration) *Ticker {
	c := make(chan time.Time)

	go func() {
		defer close(c)

		for {
			if err := Sleep(ctx, Pause(minPause, maxPause)); err != nil {
				return
			}
			c <- time.Now()
		}
	}()

	return &Ticker{C: c}
}

// Loop repeatedly waits for a random pause, then calls fn until ctx is cancelled.
// It returns nil when ctx is done.
func Loop(ctx context.Context, minPause, maxPause time.Duration, fn func(context.Context)) {
	for {
		if err := Sleep(ctx, Pause(minPause, maxPause)); err != nil {
			return
		}
		fn(ctx)
	}
}
