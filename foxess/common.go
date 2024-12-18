package foxess

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/teh-hippo/foxess-exporter/util"
)

type Config struct {
	APIKey string `short:"k" long:"api-key" description:"FoxESS API Key"      env:"API_KEY" required:"true"`
	Debug  bool   `short:"d" long:"debug"   description:"Enable debug output" env:"DEBUG"`
}

type CustomTime struct {
	time.Time
}

type NumberAsNil struct {
	Number float64
}

type ParamHolder interface {
	apiKey() string
}

func isError(errorNumber int, message string) error {
	if errorNumber != 0 {
		return fmt.Errorf("error response from foxess: %d - %s", errorNumber, message)
	}

	return nil
}

func (t *CustomTime) UnmarshalJSON(b []byte) error {
	const format string = `"2006-01-02 15:04:05 MST-0700"`

	value := string(b)

	date, err := time.Parse(format, value)
	if err != nil {
		return fmt.Errorf("failed to parse '%s' as date of format '%s': %w", value, format, err)
	}

	t.Time = date

	return nil
}

func (t *NumberAsNil) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' {
		data = data[1 : len(data)-1]
		if len(data) == 0 {
			return nil
		}
	}

	if err := json.Unmarshal(data, &t.Number); err != nil {
		return fmt.Errorf("failed to parse '%s': %w", data, err)
	}

	return nil
}

func CalculateSignature(path, apiKey string, timestamp int64) string {
	term := []byte(path + "\\r\\n" + apiKey + "\\r\\n" + strconv.FormatInt(timestamp, 10))

	return fmt.Sprintf("%x", md5.Sum(term)) //nolint:gosec
}

func (api *Config) NewRequest(operation, path string, params, result interface{}) error {
	url := "https://www.foxesscloud.com" + path
	timestamp := time.Now().UnixMilli()
	signature := CalculateSignature(path, api.APIKey, timestamp)
	operationParts := strings.Split(operation, "/")
	operationName := operationParts[int(math.Max(0, float64(len(operationParts)-1)))]

	var (
		body io.Reader
		err  error
	)

	if params != nil {
		body, err = util.ToReader(params)
		if err != nil {
			return fmt.Errorf("unable to transform params to a reader: %w", err)
		}
	}

	request, err := http.NewRequestWithContext(context.Background(), operation, url, body)
	if err != nil {
		return fmt.Errorf("failed to create %s request to %s: %w", operation, url, err)
	}

	request.Header.Set("Token", api.APIKey)
	request.Header.Set("Signature", signature)
	request.Header.Set("Timestamp", strconv.FormatInt(timestamp, 10))
	request.Header.Set("Lang", "en")
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform %s request to %s: %w", operation, url, err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read the body of %s request to %s: %w", operation, url, err)
	}

	if api.Debug {
		err = util.ToFile(fmt.Sprintf("debug-%s-%d.json", operationName, timestamp), data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing json: %v\n", err)
		}
	}

	if err = json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to unmarshal response from %s request to %s: %w", operation, url, err)
	}

	if api.Debug {
		if err = util.ToFile(fmt.Sprintf("debug-%s-%d-marshalled.json", operationName, timestamp), data); err != nil {
			// Output the error, but continue with the result.
			fmt.Fprintf(os.Stderr, "error writing json: %v\n", err)
		}
	}

	return nil
}
