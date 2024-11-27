package foxess

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func NewRequest(apiKey string, operation string, path string, params interface{}, result interface{}, debug bool) error {
	url := BaseUrl + path
	timestamp := time.Now().UnixMilli()
	signature := CalculateSignature(path, apiKey, timestamp)
	var body io.Reader
	var err error

	if params != nil {
		body, err = util.ToReader(params)
		if err != nil {
			return err
		}
	}

	request, err := http.NewRequest(operation, url, body)
	if err != nil {
		return err
	}
	request.Header.Set("token", apiKey)
	request.Header.Set("signature", signature)
	request.Header.Set("timestamp", fmt.Sprint(timestamp))
	request.Header.Set("lang", "en")
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if debug {
		err = util.ToFile(fmt.Sprintf("operation-%s-%d.json", operation, timestamp), data)
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		return err
	}
	return nil
}
