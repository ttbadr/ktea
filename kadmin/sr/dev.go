//go:build dev

package sr

import "time"

func maybeIntroduceLatency() {
	time.Sleep(1 * time.Second)
}
