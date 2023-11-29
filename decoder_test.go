package lazyhttp_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/niksteff/lazyhttp"
)

func TestDecodeToJson(t *testing.T) {
	type Data struct {
		Value string
	}

	testData := []byte(`{"value": "foo"}`)
	r := bytes.NewBuffer(testData)
	rc := io.NopCloser(r)

	decoded, err := lazyhttp.DecodeToJson[Data](rc)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if decoded.Value != "foo" {
		t.Errorf("unexpected value: %s", decoded.Value)
	}
}

func TestDecodeToBytes(t *testing.T) {
	testData := []byte(`{"value": "foo"}`)
	r := bytes.NewBuffer(testData)
	rc := io.NopCloser(r)

	decoded, err := lazyhttp.DecodeToBytes(rc)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !bytes.Equal(decoded, testData) {
		t.Errorf("unexpected value: %s expected %s", decoded, testData)
	}
}