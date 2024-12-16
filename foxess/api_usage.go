package foxess

import (
	"encoding/json"
	"fmt"
)

type AccessCountResponse struct {
	ErrorNumber int `json:"errno"`
	Result      struct {
		Total     json.Number `json:"total"`
		Remaining json.Number `json:"remaining"`
	} `json:"result"`
}

type ApiUsage struct {
	Total          float64
	Remaining      float64
	PercentageUsed float64
}

func (api *FoxessApi) GetApiUsage() (*ApiUsage, error) {
	response := &AccessCountResponse{}
	err := api.NewRequest("GET", "/op/v0/user/getAccessCount", nil, response)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest api usage: %w", err)
	}

	if err = IsError(response.ErrorNumber, ""); err != nil {
		return nil, err
	}

	total, err := response.Result.Total.Float64()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to float '%v': %w", response.Result.Total, err)
	}

	remaining, err := response.Result.Remaining.Float64()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to float '%v': %w", response.Result.Remaining, err)
	}

	percentage := (total - remaining) / total * 100
	return &ApiUsage{
		Total:          total,
		Remaining:      remaining,
		PercentageUsed: percentage,
	}, nil
}
