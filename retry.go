package lazyhttp

import "net/http"

// NoopRetryHook is a retry hook that never retries
func NoopRetryHook(resp *http.Response) bool {
	return false
}

// TODO: implement amount of retries