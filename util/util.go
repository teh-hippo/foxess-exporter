package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func ToReader(v interface{}) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %v %w", v, err)
	}

	return bytes.NewReader(data), nil
}

func Pluralise(length int) string {
	if length == 1 {
		return ""
	}

	return "s"
}

func FromFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %w", fileName, err)
	}
	defer file.Close()

	// Read the file
	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", fileName, err)
	}

	log.Printf("Read file: %s\n", fileName)

	return contents, nil
}

func JSONToStdOut(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	return enc.Encode(v)
}

func Clamp(value time.Duration, minLimit time.Duration, maxLimit time.Duration) time.Duration {
	if value < minLimit {
		return minLimit
	} else if value > maxLimit {
		return maxLimit
	}

	return value
}

func ToFile(fileName string, contents []byte) error {
	// Create or open the file
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w", fileName, err)
	}
	defer file.Close()

	// Write the string to the file
	_, err = file.Write(contents)
	if err != nil {
		return fmt.Errorf("failed to write to file '%s': %w", fileName, err)
	}

	return nil
}
