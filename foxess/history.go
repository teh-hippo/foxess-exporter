package foxess

type HistoryRequest struct {
	SerialNumber string   `json:"sn"`
	Begin        int64    `json:"begin"`
	End          int64    `json:"end"`
	Variables    []string `json:"variables"`
}

type HistoryResponse struct {
	ErrorNumber int    `json:"errno"`
	Message     string `json:"msg"`
	Result      []struct {
		Variables []VariableHistory `json:"datas"`
		DeviceSN  string            `json:"deviceSN"`
	} `json:"result"`
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

func (api *Config) GetVariableHistory(inverter string, begin, end int64, variables []string) ([]VariableHistory, error) {
	request := &HistoryRequest{
		Begin:        begin,
		End:          end,
		SerialNumber: inverter,
		Variables:    variables,
	}

	response := &HistoryResponse{}
	err := api.NewRequest("POST", "/op/v0/device/history/query", request, response)

	if err != nil {
		return nil, err
	} else if err = isError(response.ErrorNumber, response.Message); err != nil {
		return nil, err
	}

	result := make([]VariableHistory, len(response.Result))
	for i, r := range response.Result {
		result[i] = r.Variables[i]
	}

	return result, nil
}
