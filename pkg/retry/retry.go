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
	size := int(rate + 1)
	var errorTimestamps []time.Time

	return func() error {
		attempt := 0

		for {
			err := f()
			if err == nil {
				return nil
			}

			attempt++

			now := time.Now()

			errorTimestamps = append(errorTimestamps, now)

			if len(errorTimestamps) > size {
				errorTimestamps = errorTimestamps[1:]
			}
			if len(errorTimestamps) < size {
				continue
			}

			first := errorTimestamps[0]
			last := errorTimestamps[len(errorTimestamps)-1]

			if last.Sub(first) > time.Second {
				continue
			}

			currErrorRate := float32(len(errorTimestamps)) / float32(duration.Seconds())
			if shouldRetry(err, attempt) && currErrorRate < rate {
				continue
			}
			return err
		}
	}
}
