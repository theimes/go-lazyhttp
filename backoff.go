package lazyhttp

import (
	"math/rand"
	"time"
)

type Backoff interface {
	// Backoff returns a func that returns a duration and a bool. The duration
	// is the time to wait until the next retry. The bool indicates if a next
	// retry should be performed or not. The pkg provides a default set of
	// possible retries.
	Backoff() (time.Duration, bool)
}

// NoopBackoffFunc is a backoff function that stops retrying immediately.
type noopBackoffFunc struct{}

func NewNoopBackoff() *noopBackoffFunc {
	return &noopBackoffFunc{}
}

func (b *noopBackoffFunc) Backoff() (time.Duration, bool) {
	return 0, false
}

type infiniteBackoff struct {
	base time.Duration
}

func NewInfiniteBackoff(base time.Duration) *infiniteBackoff {
	return &infiniteBackoff{
		base: base,
	}
}

// infiniteBackoff returns an infinite backoff function. The function returned, if
// called multiple times, will always return the base duration. It will never
// return false to indicate that the max retries is reached.
func (b *infiniteBackoff) Backoff() (time.Duration, bool) {
	return b.base, true
}

type limitedTriesBackoff struct {
	base    time.Duration
	done    int
	retries int
}

func NewLimitedTriesBackoff(base time.Duration, retries int) *limitedTriesBackoff {
	return &limitedTriesBackoff{
		base:    base,
		done:    0,
		retries: retries,
	}
}

// limitedTriesBackoff returns a limited tries backoff function. The function
// returned, if called multiple times, will always return the base duration and
// after the max retries is reached it will return false.
func (b *limitedTriesBackoff) Backoff() (time.Duration, bool) {
	if b.done+1 > b.retries {
		return 0, false
	}

	b.done++

	return b.base, true

}

type exponentialBackoff struct {
	base    time.Duration
	done    int
	max     time.Duration
	retries int
	jitter  func() time.Duration
}

func NewExponentialBackoff(base time.Duration, max time.Duration, retries int) *exponentialBackoff {
	b := &exponentialBackoff{
		base:    base,
		max:     max,
		retries: retries,
	}

	// smallest interval is 1 second
	if b.base == 0 {
		// we really do not want anyone to retry http requests instantly with
		// an exponential backoff method. If you forget to set this value to
		// anything bigger than 0, a panic will occure to remind you to set it.
		panic("exponential retry base duration must be greater than 0")
	}

	// init a func that calculates a jitter to prevent a thundering herd
	b.jitter = func() time.Duration {
		return time.Duration(rand.Intn(500)) * time.Millisecond
	}

	return b
}

// expBackoff returns am exponential backoff function. The function returned, if
// called multiple times, will exponentially increase the returned duration
// beginning at the base duration until it reaches the max duration. After this
// it returns the max duration. Returns false when the max count of retries is
// reached.
func (b *exponentialBackoff) Backoff() (time.Duration, bool) {
	// if we reached max retries, we return false to indicate we are done
	if b.done+1 > b.retries {
		return 0, false
	}

	var current time.Duration
	if b.done == 0 {
		current = b.base + b.jitter()
	} else {
		current = (b.base * time.Duration(2<<(b.done-1))) + b.jitter()
	}

	if current >= b.max {
		b.done++
		return b.max + b.jitter(), true
	}

	b.done++
	return current, true
}
