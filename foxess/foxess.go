package foxess

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/teh-hippo/foxess-exporter/util"
)

const (
	BaseUrl = "https://www.foxesscloud.com"
)

type Foxess struct {
	ApiKey   string
	Inverter string
}

type HistoryRequest struct {
	SerialNumber string `json:"sn"`
}

func CalculateSignature(path string, apiKey string, timestamp int64) string {
	term := []byte(path + "\\r\\n" + apiKey + "\\r\\n" + fmt.Sprint(timestamp))
	return fmt.Sprintf("%x", md5.Sum(term))
}

func (g *Foxess) GetHistory(start time.Time, end time.Time) error {
	const (
		path = "/op/v0/device/history/query"
		url  = BaseUrl + path
	)
	timestamp := time.Now().UnixMilli()
	signature := CalculateSignature(path, g.ApiKey, timestamp)
	params := &HistoryRequest{SerialNumber: g.Inverter}
	body, err := util.ToReader(params)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	request.Header.Set("token", g.ApiKey)
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
