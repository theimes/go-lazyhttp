package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type NoTokenError struct {
	Err error
}

func (e NoTokenError) Error() string {
	return fmt.Sprintf("no token available in time: %s", e.Err.Error())
}

type tokenBucketRateLimiter struct {
	t       *time.Ticker  // tell us how often to fill the bucket
	timeout time.Duration // the maximum time to wait for a token if the bucket is empty and the caller does not provide a deadline

	mtx    *sync.Mutex   // protect the bucket to allow concurrent access
	bucket chan struct{} // the bucket
}

func NewTokenBucketRateLimiter(t time.Ticker, maxTokens int, timeout time.Duration) *tokenBucketRateLimiter {
	if timeout == 0 {
		timeout = time.Second * 30
	}

	lim := &tokenBucketRateLimiter{
		t:       &t,
		timeout: timeout,
		mtx:     &sync.Mutex{},
		bucket:  make(chan struct{}, maxTokens),
	}

	go func() {
		// prefill the bucket
		for i := 0; i < maxTokens; i++ {
			lim.bucket <- struct{}{}
		}

		// for each tick fill up the bucket
		for range t.C {
			lim.mtx.Lock()

			// fill the bucket
			for i := 0; i < maxTokens-len(lim.bucket); i++ {
				lim.bucket <- struct{}{}
			}

			lim.mtx.Unlock()
		}
	}()

	return lim
}

func (l *tokenBucketRateLimiter) Wait(ctx context.Context) error {
	_, ok := ctx.Deadline()
	if !ok {
		// predeclare cancel so we can wrap the parent ctx in the scope of the
		// if statement
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, l.timeout)
		defer cancel()
	}

	select {
	case <-ctx.Done():
		// context timed out, return an error describing that no token was
		// available in the allowed time frame.
		return NoTokenError{
			Err: ctx.Err(),
		}
	case <-l.bucket:
		return nil
	}
}
