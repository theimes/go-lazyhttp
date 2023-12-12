package lazyhttp_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/niksteff/lazyhttp"
)

const (
	B  = 1
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

func TestDecodeToBytes(t *testing.T) {
	testData := []byte(`{"value": "foo"}`)
	r := bytes.NewBuffer(testData)
	rc := io.NopCloser(r)

	decoded, err := lazyhttp.DecodeBytes(rc)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !bytes.Equal(decoded, testData) {
		t.Errorf("unexpected value: %s expected %s", decoded, testData)
	}
}

func TestDecodeJson(t *testing.T) {
	d := `{
		"foo": "bar"
	}`

	type Res struct {
		Foo string `json:"foo"`
	}

	var res Res
	err := lazyhttp.DecodeJson(io.NopCloser(strings.NewReader(d)), &res)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Logf("result: %#v", res)
}

func TestDecodeJsonLimit(t *testing.T) {
	d := `{
		"foo": "bar"
	}`

	type Res struct {
		Foo string `json:"foo"`
	}

	var res Res
	err := lazyhttp.DecodeJson(io.NopCloser(io.LimitReader(strings.NewReader(d), 1*KB)), &res)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	t.Logf("result: %#v", res)
}

func TestDecodeJsonLimitTooLong(t *testing.T) {
	d := `{
		"foo": "bar"
	}`

	type Res struct {
		Foo string `json:"foo"`
	}

	var res Res
	err := lazyhttp.DecodeJson(io.NopCloser(io.LimitReader(strings.NewReader(d), 1*B)), &res)
	if err != nil {
		if res != (Res{}) {
			t.Errorf("expected empty result but got: %#v", res)
		}
	}
}
