//go:build dev

package kadmin

import "time"

func MaybeIntroduceLatency() {
	time.Sleep(2 * time.Second)
}
