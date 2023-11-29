package lazyhttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

type RequestOption func(r *http.Request) error

func WithContext(ctx context.Context) RequestOption {
	return func(r *http.Request) error {
		if ctx == nil {
			return RequestError{
				Err:     fmt.Errorf("context is nil"),
				Request: r,
			}
		}

		// set the context for the request
		_ = r.WithContext(ctx)

		return nil
	}
}

func WithHeaders(h http.Header) RequestOption {
	return func(r *http.Request) error {
		r.Header = h
		return nil
	}
}

func WithBody(body io.Reader) RequestOption {
	return func(r *http.Request) error {
		// first we got to convert the input to a read closer so we can assign
		// it to our request body
		rc, ok := body.(io.ReadCloser)
		if !ok && body != nil {
			// use a noop closer if the body is not a read closer to fullfil
			// the required interface
			rc = io.NopCloser(body)
		}

		r.Body = rc

		return nil
	}
}

func WithBytesBody(body []byte) RequestOption {
	return WithBody(bytes.NewReader(body))
}

func WithBasicAuth(username, password string) RequestOption {
	return func(r *http.Request) error {
		r.SetBasicAuth(username, password)
		return nil
	}
}

// NewRequest calls NewRequestWithContext with a background context.
func NewRequest(method string, url string, opts ...RequestOption) (*http.Request, error) {
	return NewRequestWithContext(context.Background(), method, url, opts...)
}

// NewRequestWithContext creates a new http request with the given context and
// provides the option pattern to further change this request.
func NewRequestWithContext(ctx context.Context, method string, url string, opts ...RequestOption) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, RequestError{
			Err: fmt.Errorf("error creating request: %w", err),
		}
	}

	for _, opt := range opts {
		err := opt(req)
		if err != nil {
			return nil, RequestError{
				Err:     fmt.Errorf("error applying request option: %w", err),
				Request: req,
			}
		}
	}

	return req, nil
}
