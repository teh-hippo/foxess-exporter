package foxess

type RealTimeRequest struct {
	SerialNumbers []string `json:"sns"`
	Variables     []string `json:"variables"`
}

type RealTimeResponse struct {
	ErrorNumber int            `json:"errno"`
	Message     string         `json:"msg"`
	Result      []RealTimeData `json:"result"`
}

type RealTimeData struct {
	Variables []struct {
		Variable string      `json:"variable"`
		Unit     string      `json:"unit"`
		Name     string      `json:"name"`
		Value    NumberAsNil `json:"value"`
	} `json:"datas"`
	DeviceSN string     `json:"deviceSN"`
	Time     CustomTime `json:"time"`
}

func (api *FoxessAPI) GetRealTimeData(inverters []string, variables []string) ([]RealTimeData, error) {
	request := &RealTimeRequest{
		SerialNumbers: inverters,
	}
	if variables != nil {
		request.Variables = variables
	}
	response := &RealTimeResponse{}
	if err := api.NewRequest("POST", "/op/v1/device/real/query", request, response); err != nil {
		return nil, err
	} else if err = isError(response.ErrorNumber, response.Message); err != nil {
		return nil, err
	} else {
		return response.Result, nil
	}
}
