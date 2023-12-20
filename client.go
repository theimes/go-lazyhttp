// lazyhttp is a simple wrapper around the go http client that provides various
// convencience functions to implement common http client usecases in go. The
// library is completely compatible with the go http pkg. It uses the functional
// options pattern to configure the client and requests. The client provides
// wrapped errors and custom error types for you to check which step of a
// request went fubar.
//
// What lazyhttp is and is not:
// - It is NOT an opinionated http client.
// - It provides sensible defaults.
// - You can always manipulate your request and response objects in any way you
//   want to because you have direct access to these.
// - It provides custom error types for error identification with errors.Is()
//   and errors.As().
//
// Why should I use it instead of using the go standard library?
// - You should not. If you are happy with the go standard library you should
//   always prefer the go standard lib. This library is for people who want to
//   have a little bit more convenience and do not or not yet know the internals
//   of the go http client.

package lazyhttp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var ErrMaxRetriesReached error = fmt.Errorf("max retries reached")

// Authenticator is an interface that can be implemented to authenticate a given
// request. If the request could not be authenticated an error is returend.
type Authenticator interface {
	Authenticate(*http.Request) error
}

// AuthenticatorFunc is a function that implements the Authenticator interface
type AuthenticatorFunc func(*http.Request) error

// Authenticate calls the Authentifactor function to implement the interface
func (f AuthenticatorFunc) Authenticate(r *http.Request) error {
	return f(r)
}

type RateLimiter interface {
	Wait(ctx context.Context) error
}

// Option implements the functional options pattern for the client
type Option func(*client) *client

// PreRequestHook is a function that is called before the request is made. It
// can alter the request before the request is made.
type PreRequestHook func(*http.Request) error

// RetryPolicy is a function that is called after the response is received. It
// decides if the request should be retried or not.
type RetryPolicy func(*http.Response) bool

// PostResponseHook is a function that is called after the response is received.
// It can alter the response before it is returned.
type PostResponseHook func(*http.Response) error

type Config struct {
	MaxRateLimiterWaitTime time.Duration
}

type client struct {
	conf          Config
	httpClient    *http.Client       // the underlying http client, this can be configured
	rateLimiter   RateLimiter        // the rate limiter, this can be configured
	preReqHooks   []PreRequestHook   // functions that are ran before the request is made
	retryPolicy   RetryPolicy        // function that is ran after the response is received to decide if the request should be retried
	newBackoffPolicy func() Backoff     // a function that returns a new instance of a backoff implementation
	postRespHooks []PostResponseHook // functions that are ran after the response is received
	authenticator Authenticator      // authenticator that is used to authenticate each request
	host          *url.URL           // the host url that is used for all requests
}

func WithHttpClient(httpClient *http.Client) Option {
	return func(c *client) *client {
		c.httpClient = httpClient
		return c
	}
}

func WithRateLimiter(rateLimiter RateLimiter) Option {
	return func(c *client) *client {
		c.rateLimiter = rateLimiter
		return c
	}
}

func WithMaxRateLimiterWaitTime(d time.Duration) Option {
	return func(c *client) *client {
		c.conf.MaxRateLimiterWaitTime = d
		return c
	}
}

func WithPreRequestHooks(hook ...PreRequestHook) Option {
	return func(c *client) *client {
		c.preReqHooks = append(c.preReqHooks, hook...)
		return c
	}
}

func WithPostResponseHooks(hook ...PostResponseHook) Option {
	return func(c *client) *client {
		c.postRespHooks = append(c.postRespHooks, hook...)
		return c
	}
}

func WithAuthenticator(authenticator Authenticator) Option {
	return func(c *client) *client {
		c.authenticator = authenticator
		return c
	}
}

func WithHost(host *url.URL) Option {
	return func(c *client) *client {
		c.host = host
		return c
	}
}

// WithRetryPolicy sets a function that is called after the response is received
// it decides whether to retry the request based on the response. You have to
// implement this hook yourself. The pkg provides a basic NoopRetryHook that
// will never perform a retry.
func WithRetryPolicy(hook RetryPolicy) Option {
	return func(c *client) *client {
		c.retryPolicy = hook
		return c
	}
}

// WithBackoffPolicy sets a function that returns a new instance of a `Backoff`
// implementation on each new request. The pkg provides basic backoff mechanisms
// you can use. If you want to implement your own backoff mechanism you can do
// so by implementing the `Backoff` interface yourself.
func WithBackoffPolicy(backoff func() Backoff) Option {
	return func(c *client) *client {
		c.newBackoffPolicy = backoff
		return c
	}
}

// NewClient creates a new client with the given options. If no options are
// given sensible defaults are selected.
func NewClient(opts ...Option) *client {
	httpClient := http.DefaultClient      // go's default http client is also our default
	httpClient.Timeout = 30 * time.Second // this is our sensible default for timeouts

	c := &client{
		conf: Config{
			MaxRateLimiterWaitTime: 60 * time.Second,
		},
		httpClient:    httpClient,
		rateLimiter:   nil,                                        // no default rate limiter
		preReqHooks:   []PreRequestHook{},                         // no default pre request hooks
		retryPolicy:   nil,                                        // by default never retry anything
		newBackoffPolicy: func() Backoff { return NewNoopBackoff() }, // default backoff implementation is an exponential backoff with defensive values
		postRespHooks: []PostResponseHook{},                       // no default post response hooks
		authenticator: nil,                                        // no default authenticator
	}

	// apply the given options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	// first the get the context from the request so we operate on the same
	ctx := req.Context()

	// if a rate limiter is set we got to wait for allowance. Run the
	// ratelimiter before everything else because of there is no free token we
	// do not bother.
	if c.rateLimiter != nil {
		// if the given context has no deadline we se the default deadline from
		// the client to protect the user from never ending waits.
		_, ok := req.Context().Deadline()
		if !ok {
			// wrap the request context with a deadline to provide a timeout
			// for rate limit waits even if the user does not provide a deadline in
			// the context.
			var cancel context.CancelFunc
			ctx, cancel = context.WithDeadline(req.Context(), time.Now().Add(c.conf.MaxRateLimiterWaitTime))
			defer cancel()
		}

		err := c.rateLimiter.Wait(ctx)
		if err != nil {
			return nil, RateLimitError{
				Err:         err,
				RateLimiter: c.rateLimiter,
			}
		}
	}

	// run all the pre request hooks
	if c.preReqHooks != nil {
		for _, hook := range c.preReqHooks {
			err := hook(req)
			if err != nil {
				return &http.Response{}, RequestError{
					Err:     fmt.Errorf("error running pre request hook: %w", err),
					Request: req,
				}
			}
		}
	}

	// authenticate the request
	if c.authenticator != nil {
		err := c.authenticator.Authenticate(req)
		if err != nil {
			return nil, RequestError{
				Err:     fmt.Errorf("error authenticating request: %w", err),
				Request: req,
			}
		}
	}

	// set the host
	if c.host != nil {
		if req.URL.Scheme == "" {
			req.URL.Scheme = c.host.Scheme
		}

		if req.URL.Host == "" {
			req.URL.Host = c.host.Host
		}
	}

	// now execute the request
	res, err := c.execute(req)
	if err != nil {
		return res, fmt.Errorf("error executing request: %w", err)
	}

	// handle all retry operations
	if c.retryPolicy != nil {
		// create a new backoff instance for this request
		bop := c.newBackoffPolicy()

		// check if the retry hook wants to perform a retry
		for c.retryPolicy(res) {
			var err error

			// want to perform a retry so check the backoff implementation if
			// a retry is still possible
			t, ok := bop.Backoff()
			if !ok {
				return res, fmt.Errorf("error backing off from request: %w", err)
			}

			// wait for the backoff deadline
			timer := time.NewTimer(t)

			// run a func that returns as soon as either the context is done or
			// we received a response from the request.
			res, err = func(res *http.Response) (*http.Response, error) {
				for {
					select {
					case <-ctx.Done():
						return res, fmt.Errorf("request context done")
					case <-timer.C:
						// now execute the request without all prior hooks etc.
						// because we already did that.
						res, err = c.execute(req)
						if err != nil {
							return res, fmt.Errorf("error executing request: %w", err)
						}
						// TODO: check this timer behaviour
						timer.Stop()
					
						return res, nil
					}
				}
			}(res)
		}
	}

	// run all the post response hooks
	if c.postRespHooks != nil {
		for _, hook := range c.postRespHooks {
			err := hook(res)
			if err != nil {
				return res, ResponseError{
					Err:      fmt.Errorf("error running post response hook: %w", err),
					Response: res,
				}
			}
		}
	}

	return res, nil
}

func (c *client) execute(r *http.Request) (*http.Response, error) {
	// finally perform the request
	res, err := c.httpClient.Do(r)
	if err != nil {
		return nil, RequestError{
			Err:     fmt.Errorf("error making http request: %w", err),
			Request: r,
		}
	}

	return res, nil
}
