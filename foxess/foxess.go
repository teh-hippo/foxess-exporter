package foxess

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/teh-hippo/foxess-exporter/util"
)

const (
	BaseUrl = "https://www.foxesscloud.com"
)

func CalculateSignature(path string, apiKey string, timestamp int64) string {
	term := []byte(path + "\\r\\n" + apiKey + "\\r\\n" + fmt.Sprint(timestamp))
	return fmt.Sprintf("%x", md5.Sum(term))
}

func (api *FoxessApi) NewRequest(operation string, path string, params interface{}, result interface{}) error {
	url := BaseUrl + path
	timestamp := time.Now().UnixMilli()
	signature := CalculateSignature(path, api.ApiKey, timestamp)
	operationParts := strings.Split(operation, "/")
	operationName := operationParts[int(math.Max(0, float64(len(operationParts)-1)))]
	var body io.Reader
	var err error

	if params != nil {
		body, err = util.ToReader(params)
		if err != nil {
			return fmt.Errorf("unable to transform params to a reader: %w", err)
		}
	}

	request, err := http.NewRequest(operation, url, body)
	if err != nil {
		return fmt.Errorf("failed to create %s request to %s: %w", operation, url, err)
	}

	request.Header.Set("token", api.ApiKey)
	request.Header.Set("signature", signature)
	request.Header.Set("timestamp", fmt.Sprint(timestamp))
	request.Header.Set("lang", "en")
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform %s request to %s: %w", operation, url, err)
	}

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
