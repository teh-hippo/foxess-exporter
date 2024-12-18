package foxess

import "fmt"

const PageSize = 1000

type DeviceListRequest struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
}

const (
	StatusOnline = iota + 1
	StatusFault
	StatusOffline
)

type DeviceListResponse struct {
	ErrorNumber int `json:"errno"`
	Result      struct {
		CurrentPage int      `json:"currentPage"`
		PageSize    int      `json:"pageSize"`
		Total       int      `json:"total"`
		Devices     []Device `json:"data"`
	}
}

type Device struct {
	DeviceSerialNumber string `json:"deviceSN"`
	ModuleSerialNumber string `json:"moduleSN"`
	StationID          string `json:"stationId"`
	StationName        string `json:"stationName"`
	Status             int    `json:"status"`
	HasPV              bool   `json:"hasPV"`
	HasBattery         bool   `json:"hasBattery"`
	DeviceType         string `json:"deviceType"`
	ProductType        string `json:"productType"`
}

func (api *Config) GetDeviceList() ([]Device, error) {
	currentPage := 1
	total := 1
	devices := make([]Device, 0)

	for len(devices) < total {
		request := &DeviceListRequest{
			CurrentPage: currentPage,
			PageSize:    PageSize,
		}
		response := &DeviceListResponse{} //nolint:exhaustruct

		if err := api.NewRequest("POST", "/op/v0/device/list", request, response); err != nil {
			return nil, err
		} else if err = isError(response.ErrorNumber, ""); err != nil {
			return nil, err
		}

		devices = append(devices, response.Result.Devices...)
		total = response.Result.Total
		currentPage++
	}

	return devices, nil
}

func (d *Device) CurrentStatus() string {
	switch d.Status {
	case StatusOnline:
		return "Online"
	case StatusFault:
		return "Fault"
	case StatusOffline:
		return "Offline"
	default:
		return fmt.Sprint("Unknown:", d.Status)
	}
}
