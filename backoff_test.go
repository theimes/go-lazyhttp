package lazyhttp_test

import (
	"testing"
	"time"

	"github.com/niksteff/go-lazyhttp"
)

func TestBackoffIncreases(t *testing.T) {
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

func TestBackoffConsidersMax(t *testing.T) {
	retries := 5
	b := lazyhttp.NewExponentialBackoff(5*time.Second, 2*time.Second, retries)

	// Test if the backoff time is capped by the max duration
	for i := 0; i < retries; i++ {
		next, _ := b.Backoff()
		if next > 2500*time.Millisecond { // here we know, that the jitter is smaller than 500ms
			t.Errorf("Backoff time exceeded max: next = %v", next)
		}
	}
}
