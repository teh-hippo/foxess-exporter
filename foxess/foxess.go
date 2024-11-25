package foxess

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	BaseUrl = "https://www.foxesscloud.com"
)

func CalculateSignature(path string, apiKey string, timestamp int64) string {
	term := []byte(path + "\\r\\n" + apiKey + "\\r\\n" + fmt.Sprint(timestamp))
	return fmt.Sprintf("%x", md5.Sum(term))
}

func GetHistory(apiKey string, inverter string) error {
	const (
		path = "/op/v0/device/history/query"
		url  = BaseUrl + path
	)
	timestamp := time.Now().UnixMilli()
	signature := CalculateSignature(path, apiKey, timestamp)
	body := "{\n\"sn\":\"" + inverter + "\"\n}"
	request, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("token", apiKey)
	request.Header.Set("signature", signature)
	request.Header.Set("timestamp", fmt.Sprint(timestamp))
	request.Header.Set("lang", "en")
	request.Header.Set("Content-Type", "application/json")
	log.Printf("Request: %v\n", request)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(result))
	return nil
}
