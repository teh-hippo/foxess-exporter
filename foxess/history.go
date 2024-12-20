package foxess

import (
	"sort"
	"time"
)

type HistoryRequest struct {
	SerialNumber string   `json:"sn"`
	Begin        int64    `json:"begin"`
	End          int64    `json:"end"`
	Variables    []string `json:"variables"`
}

type InverterHistory struct {
	Variables []VariableHistory `json:"datas"`
	DeviceSN  string            `json:"deviceSN"`
}

type HistoryResponse struct {
	ErrorNumber int               `json:"errno"`
	Message     string            `json:"msg"`
	Result      []InverterHistory `json:"result"`
}

type DataPoint struct {
	Time  CustomTime  `json:"time"`
	Value NumberAsNil `json:"value"`
}

type VariableHistory struct {
	Unit       string      `json:"unit"`
	DataPoints []DataPoint `json:"data"`
	Name       string      `json:"name"`
	Variable   string      `json:"variable"`
}

func (api *Config) GetVariableHistory(inverter string, begin, end time.Time, variables []string) ([]InverterHistory, error) {
	request := &HistoryRequest{
		Begin:        begin.UnixMilli(),
		End:          end.UnixMilli(),
		SerialNumber: inverter,
		Variables:    variables,
	}

	response := &HistoryResponse{} //nolint:exhaustruct
	err := api.NewRequest("POST", "/op/v0/device/history/query", request, response)

	if err != nil {
		return nil, err
	} else if err = isError(response.ErrorNumber, response.Message); err != nil {
		return nil, err
	}

	for i := range response.Result {
		for _, r := range response.Result[i].Variables {
			sort.Slice(r.DataPoints, func(i, j int) bool {
				return r.DataPoints[i].Time.UnixMilli() < r.DataPoints[j].Time.UnixMilli()
			})
		}
	}

	return response.Result, nil
}
