//go:build dev

package sradmin

import "time"

func maybeIntroduceLatency() {
	time.Sleep(1 * time.Second)
}
