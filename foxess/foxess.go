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

type Foxess struct {
	ApiKey   string
	Inverter string
	Debug    bool
}

type HistoryRequest struct {
	SerialNumber string `json:"sn"`
}

type Variable struct {
	Unit                  string `json:"unit"`
	GridTiedInverter      bool   `json:"Grid-tied inverter"`
	EnergyStorageInverter bool   `json:"Energy-storage inverter"`
}

// Define the structure for the response
type VariablesResponse struct {
	Errno  int                   `json:"errno"`
	Msg    string                `json:"msg"`
	Result []map[string]Variable `json:"result"`
}

func CalculateSignature(path string, apiKey string, timestamp int64) string {
	term := []byte(path + "\\r\\n" + apiKey + "\\r\\n" + fmt.Sprint(timestamp))
	return fmt.Sprintf("%x", md5.Sum(term))
}

func (g *Foxess) GetHistory(start time.Time, end time.Time) error {
	return g.Request("POST", "/op/v0/device/history/query", &HistoryRequest{SerialNumber: g.Inverter}, &HistoryRequest{})
}

func (g *Foxess) GetAvailableVariables() (*VariablesResponse, error) {
	result := &VariablesResponse{}
	err := g.Request("GET", "/op/v0/device/variable/get", nil, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *Foxess) Request(operation string, path string, params interface{}, result interface{}) error {
	url := BaseUrl + path
	timestamp := time.Now().UnixMilli()
	signature := CalculateSignature(path, g.ApiKey, timestamp)
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
	request.Header.Set("token", g.ApiKey)
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

	if g.Debug {
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
