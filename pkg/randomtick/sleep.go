package randomtick

import (
	"context"
	"time"
)

// Sleep is a context-aware sleep function that stops when ctx is cancelled.
func Sleep(ctx context.Context, duration time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
