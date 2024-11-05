//go:build dev

package kadmin

import "time"

func maybeIntroduceLatency() {
	time.Sleep(1 * time.Second)
}
