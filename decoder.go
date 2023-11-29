package lazyhttp

import (
	"encoding/json"
	"fmt"
	"io"
)

func DecodeToBytes(rc io.ReadCloser) ([]byte, error) {
	defer rc.Close()

	b, err := io.ReadAll(rc)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading response body: %w", err)
	}

	return b, nil
}

func DecodeToJson[T any](rc io.ReadCloser) (T, error) {
	defer rc.Close()

	var target T

	b, err := io.ReadAll(rc)
	if err != nil {
		return target, fmt.Errorf("error reading response body: %w", err)
	}

	err = json.Unmarshal(b, &target)
	if err != nil {
		return target, fmt.Errorf("error deserializing response body: %w", err)
	}

	return target, nil
}
