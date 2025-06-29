//go:build dev

package kadmin

import "time"

func MaybeIntroduceLatency() {
	time.Sleep(5 * time.Second)
}
