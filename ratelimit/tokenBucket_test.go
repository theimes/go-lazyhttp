package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewTokenBucketRateLimiter(t *testing.T) {
	ticker := time.NewTicker(1 * time.Millisecond)
	maxTokens := 5
	timeout := 10 * time.Millisecond

	limiter := NewTokenBucketRateLimiter(*ticker, maxTokens, timeout)
	// give a little slack for the goroutine to finish
	time.Sleep(5 * time.Millisecond)

	// Test that the limiter is created correctly
	if limiter.timeout != timeout {
		t.Errorf("Expected timeout to be %v, but got %v", timeout, limiter.timeout)
	}
	if len(limiter.bucket) != maxTokens {
		t.Errorf("Expected bucket size to be %d after creation, but got %d", maxTokens, len(limiter.bucket))
	}

	// Test it waits for a token to be available
	for i := 0; i < maxTokens; i++ {
		// emtpy the bucket
		<-limiter.bucket
	}
	// ensure the bucket is empty
	if len(limiter.bucket) == maxTokens {
		t.Errorf("Expected bucket size to be reduced after emptying, but got %d", len(limiter.bucket))
	}

	// Test that the bucket is refilled correctly
	time.Sleep(6 * time.Millisecond) // Wait for the bucket to be refilled
	if len(limiter.bucket) != maxTokens {
		t.Errorf("Expected bucket size to be %d after refill, but got %d", maxTokens, len(limiter.bucket))
	}

	// Test that the bucket is not overfilled
	time.Sleep(10 * time.Millisecond) // Wait for multiple refills
	if len(limiter.bucket) > maxTokens {
		t.Errorf("Expected bucket size to be at most %d, but got %d", maxTokens, len(limiter.bucket))
	}

	// Test that the bucket is not underfilled
	if len(limiter.bucket) < maxTokens {
		t.Errorf("Expected bucket size to be at least %d, but got %d", maxTokens, len(limiter.bucket))
	}

	// Test that the bucket is emptied correctly
	for i := 0; i < maxTokens; i++ {
		<-limiter.bucket
	}
	if len(limiter.bucket) != 0 {
		t.Errorf("Expected bucket size to be 0 after emptying, but got %d", len(limiter.bucket))
	}
}

func TestRateLimiterWait(t *testing.T) {
	tickTime := 1 * time.Second
	ticker := time.NewTicker(tickTime)
	maxTokens := 5
	timeout := 3 * time.Second

	limiter := NewTokenBucketRateLimiter(*ticker, maxTokens, timeout)

	// Test it waits for a token to be available
	for i := 0; i < maxTokens; i++ {
		// emtpy the bucket
		<-limiter.bucket
	}
	// ensure the bucket is empty
	if len(limiter.bucket) != 0 {
		t.Errorf("Expected bucket size to be reduced after emptying, but got %d", len(limiter.bucket))
	}

	// Test the waiting functionality
	startTime := time.Now()

	err := limiter.Wait(context.Background())
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	endTime := time.Now()

	// Test that the wait function blocks for the correct amount of time
	if endTime.Sub(startTime) < tickTime {
		t.Errorf("Expected wait to block for %v, but blocked for %v", tickTime, endTime.Sub(startTime))
	}

}

func TestRateLimiterWaitThrowsError(t *testing.T) {
	tickTime := 10 * time.Second
	ticker := time.NewTicker(tickTime)
	maxTokens := 5
	timeout := 500 * time.Millisecond

	limiter := NewTokenBucketRateLimiter(*ticker, maxTokens, timeout)

	// Test it waits for a token to be available
	for i := 0; i < maxTokens; i++ {
		// emtpy the bucket
		<-limiter.bucket
	}
	// ensure the bucket is empty
	if len(limiter.bucket) != 0 {
		t.Errorf("Expected bucket size to be reduced after emptying, but got %d", len(limiter.bucket))
	}

	err := limiter.Wait(context.Background())
	if err != nil {
		// Expect NoTokenError
		var noTokenError NoTokenError
		if !errors.As(err, &noTokenError) {
			t.Errorf("Expected NoTokenError, but got %v", err)
		}
	} else {
		t.Errorf("Expected NoTokenError, but got no error")
	}
}

func TestRateLimiterWaitConsidersCtxTimeout(t *testing.T) {
	tickTime := 10 * time.Second
	ticker := time.NewTicker(tickTime)
	maxTokens := 5
	timeout := 3 * time.Second

	limiter := NewTokenBucketRateLimiter(*ticker, maxTokens, timeout)

	// Test it waits for a token to be available
	for i := 0; i < maxTokens; i++ {
		// emtpy the bucket
		<-limiter.bucket
	}
	// ensure the bucket is empty
	if len(limiter.bucket) != 0 {
		t.Errorf("Expected bucket size to be reduced after emptying, but got %d", len(limiter.bucket))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err != nil {
		// Expect NoTokenError
		var noTokenError NoTokenError
		if !errors.As(err, &noTokenError) {
			t.Errorf("Expected NoTokenError, but got %v", err)
		}
	} else {
		t.Errorf("Expected NoTokenError, but got no error")
	}

}
