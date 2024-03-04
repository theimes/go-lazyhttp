package lazyhttp_test

import (
	"testing"
	"time"

	"github.com/niksteff/lazyhttp"
)

func TestBackoff(t *testing.T) {
	retries := 5
	b := lazyhttp.NewExponentialBackoff(5*time.Second, 2*time.Hour, retries)

	// Test if the backoff time increases exponentially
	prev, _ := b.Backoff()
	t.Logf("initial offset = %s", prev)
	for i := 0; i < retries-1; i++ {
		next, _ := b.Backoff()
		t.Logf("prev = %v, next = %v", prev, next)
		if next <= prev {
			t.Errorf("Backoff time did not increase: prev = %v, next = %v", prev, next)
		}
		prev = next
	}

	// Test if the retry limit is respected
	_, ok := b.Backoff()
	if ok {
		t.Errorf("Retry limit not respected: expected = false, got = %v", ok)
	}
}
