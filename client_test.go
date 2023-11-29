package lazyhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/niksteff/lazyhttp"
)

// TestBasicRequest tests a basic request with a deadline context
func TestBasicRequest(t *testing.T) {
	done, ok := t.Deadline()
	if !ok {
		t.Errorf("no deadline set")
	}

	ctx, cancel := context.WithDeadline(context.Background(), done)
	defer cancel()

	type testResponse struct {
		Value string
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"value": "test"}`))
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}))
	defer srv.Close()

	client := lazyhttp.NewClient()

	addr, err := url.Parse(srv.URL)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	req, err := lazyhttp.NewRequestWithContext(ctx, http.MethodGet, addr.String())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	tr, err := lazyhttp.DecodeToJson[testResponse](res.Body)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if tr.Value != "test" {
		t.Errorf("unexpected response: %s", tr.Value)
	}

	t.Log(tr)
}

// TestBasicRequest tests a basic request with a deadline context
func TestWithHost(t *testing.T) {
	done, ok := t.Deadline()
	if !ok {
		done = time.Now().Add(5 * time.Second)
	}

	ctx, cancel := context.WithDeadline(context.Background(), done)
	defer cancel()

	type testResponse struct {
		Value string
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/some/path/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"value": "test"}`))
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	addr, err := url.Parse(srv.URL)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	client := lazyhttp.NewClient(lazyhttp.WithHost(addr))

	req, err := lazyhttp.NewRequestWithContext(ctx, http.MethodGet, "/some/path/")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	tr, err := lazyhttp.DecodeToJson[testResponse](res.Body)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if tr.Value != "test" {
		t.Errorf("unexpected response: %s", tr.Value)
	}

	t.Logf("%#v\n", tr)
}
