package retry

import (
	"time"
)

type fn func() error
type shouldRetry func(err error, attempt int) bool

// WrapWithRetry - wraps the given function, retries it if it fails and shouldRetry returns true. Exists if errors rate
// is above the threshold.
func WrapWithRetry(f fn, shouldRetry shouldRetry, rate float32,
) func() error {
	var errorTimestamps []time.Time

	pruneOldErrors := func(now time.Time) {
		oneSecAgo := now.Add(-1 * time.Second)
		for len(errorTimestamps) > 0 && errorTimestamps[0].Before(oneSecAgo) {
			errorTimestamps = errorTimestamps[1:]
		}
	}

	return func() error {
		attempt := 0

		for {
			err := f()
			if err == nil {
				return nil
			}

			attempt++

			now := time.Now()
			pruneOldErrors(now)
			errorTimestamps = append(errorTimestamps, now)

			currErrorRate := float32(len(errorTimestamps))

			if shouldRetry(err, attempt) && currErrorRate < rate {
				continue
			}
			return err
		}
	}
}
