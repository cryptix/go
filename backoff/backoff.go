// package backoff implemenets policies for waiting differente ammounts of times when retrying something.
//
// Taken from Peter Teichman's post on http://blog.gopheracademy.com/advent-2014/backoff/
package backoff

import (
	"math/rand"
	"time"
)

type Backoff interface {
	Duration(n int) time.Duration
}

var (
	// Default is a backoff policy ranging up to 5 seconds.
	Default = IncreasePolicy{
		[]int{0, 10, 10, 100, 100, 500, 500, 3000, 3000, 5000},
	}

	Random = RandomPolacy{10 * time.Second}
)

// IncreasePolicy implements a backoff policy, randomizing its delays
// and saturating at the final value in Millis.
type IncreasePolicy struct {
	Millis []int
}

// Duration returns the time duration of the n'th wait cycle in a
// backoff policy. This is b.Millis[n], randomized to avoid thundering
// herds.
func (b IncreasePolicy) Duration(n int) time.Duration {
	if n >= len(b.Millis) {
		n = len(b.Millis) - 1
	}

	return time.Duration(jitter(b.Millis[n])) * time.Millisecond
}

// jitter returns a random integer uniformly distributed in the range
// [0.5 * millis .. 1.5 * millis]
func jitter(millis int) int {
	if millis == 0 {
		return 0
	}

	return millis/2 + rand.Intn(millis)
}

type RandomPolacy struct {
	max time.Duration
}

func (b RandomPolacy) Duration(n int) time.Duration {
	return time.Duration(rand.Int63() % int64(b.max))
}
