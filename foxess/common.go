package foxess

import (
	"fmt"
	"time"
)

func IsError(errorNumber int, message string) error {
	if errorNumber != 0 {
		return fmt.Errorf("error response from foxess: %d - %s", errorNumber, message)
	}
	return nil
}

type FoxessParams struct {
	ApiKey string `short:"k" long:"api-key" description:"FoxESS API Key" required:"true" env:"API_KEY"`
	Debug  bool   `short:"d" long:"debug" description:"Enable debug output"`
}

type CustomTime struct {
	time.Time
}

type ParamHolder interface {
	apiKey() string
}

func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {
	value := string(b)
	const format string = `"2006-01-02 15:04:05 MST-0700"`
	date, err := time.Parse(format, value)
	if err != nil {
		return fmt.Errorf("failed to parse '%s' as date of format '%s': %w", value, format, err)
	}
	t.Time = date
	return
}
