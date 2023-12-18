package lazyhttp

import (
	"fmt"
	"net/http"
)

// RequestError is an error that occurs during a built up of a request. This
// is not an error that results while performing a request.
type RequestError struct {
	Err     error
	Request *http.Request
}

func (e RequestError) Error() string {
	return fmt.Sprintf("error making request: %s", e.Err.Error())
}

type RateLimitError struct {
	Err         error
	RateLimiter RateLimiter
}

func (e RateLimitError) Error() string {
	return fmt.Sprintf("rate limit error: %s", e.Err.Error())
}

type ResponseError struct {
	Err      error
	Response *http.Response
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("error handling response: %s", e.Err.Error())
}
