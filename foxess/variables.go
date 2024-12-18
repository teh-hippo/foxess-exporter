package foxess

import "maps"

// Define the structure for the response.
type VariablesResponse struct {
	ErrorNumber int                   `json:"errno"`
	Message     string                `json:"msg"`
	Result      []map[string]Variable `json:"result"`
}

type Variable struct {
	Unit                  string `json:"unit"`
	GridTiedInverter      bool   `json:"Grid-tied inverter"`
	EnergyStorageInverter bool   `json:"Energy-storage inverter"`
}

func (api *FoxessAPI) GetVariables(gridOnly bool) (*[]map[string]Variable, error) {
	response := &VariablesResponse{}
	if err := api.NewRequest("GET", "/op/v0/device/variable/get", nil, response); err != nil {
		return nil, err
	} else if err = isError(response.ErrorNumber, response.Message); err != nil {
		return nil, err
	} else if !gridOnly {
		return &response.Result, nil
	} else {
		gridOnlyVariables := make([]map[string]Variable, 0)
		for _, variable := range response.Result {
			for key := range maps.Keys(variable) {
				item := variable[key]
				if item.GridTiedInverter {
					gridOnlyVariables = append(gridOnlyVariables, variable)
				}
			}
		}
		return &gridOnlyVariables, nil
	}
}
