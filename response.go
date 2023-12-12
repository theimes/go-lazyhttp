package lazyhttp

import (
	"encoding/json"
	"fmt"
	"io"
)

// DecodeBytes reads from the given reader and returns the content as a []byte.
// To limit the number of bytes read, use the io.LimitReader type to wrap the
// reader.
func DecodeBytes(rc io.ReadCloser) ([]byte, error) {
	// always close reader after reading
	defer rc.Close()

	b, err := io.ReadAll(rc)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading response body: %w", err)
	}

	return b, nil
}

// DecodeJson reads from the given reader and unmarshals the content into the given
// pointer. The reader is closed after reading. This function does not limit the
// number of bytes read from the reader. To limit the number of bytes read, use
// the io.LimitReader type to wrap the reader.
func DecodeJson[T any](r io.ReadCloser, out *T) error {
	// always close reader after reading
	defer r.Close()

	// read all from the given reader
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	err = json.Unmarshal(b, out)
	if err != nil {
		return fmt.Errorf("error unmarshaling response body: %w", err)
	}

	return nil
}