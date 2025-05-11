package retry

import (
	"time"
)

type fn func() error
type shouldRetry func(err error, attempt int) bool

// WrapWithRetry - wraps the given function, retries it if it fails and shouldRetry returns true. Exists if errors rate
// is above the threshold.
func WrapWithRetry(f fn, shouldRetry shouldRetry, rate float32) func() error {
	var errorTimestamps []time.Time

	return func() error {
		attempt := 0

		for {
			err := f()
			if err == nil {
				return nil
			}

			attempt++

			errorTimestamps = append(errorTimestamps, time.Now())

			if len(errorTimestamps) < int(rate+1) {
				continue
			}

			errorTimestamps = errorTimestamps[1:]

			if errorTimestamps[len(errorTimestamps)-1].Sub(errorTimestamps[0]) < time.Second ||
				!shouldRetry(err, attempt) {
				return err // TODO: join and return last N errors
			}
		}
	}
}
