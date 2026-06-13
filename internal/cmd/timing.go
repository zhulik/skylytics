package cmd

import (
	"crypto/rand"
	"math/big"
	"time"
)

func randomPause(minPause, maxPause time.Duration) time.Duration {
	span := big.NewInt(int64(maxPause - minPause + 1))
	n, err := rand.Int(rand.Reader, span)
	if err != nil {
		return minPause
	}
	return minPause + time.Duration(n.Int64())
}
