package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func ToReader(v interface{}) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %w", err)
	}
	return bytes.NewReader(data), nil
}
