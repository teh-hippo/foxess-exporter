package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func ToReader(v interface{}) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %w", err)
	}
	return bytes.NewReader(data), nil
}

func FromFile(fileName string) ([]byte, error) {
	// Open the file
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

func JsonToStdOut(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
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

	log.Printf("Wrote file: %s\n", fileName)
	return nil
}
