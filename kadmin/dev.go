//go:build dev

package kadmin

import "time"

func maybeIntroduceLatency() {
	time.Sleep(2 * time.Second)
}
