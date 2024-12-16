package foxess

import "fmt"

const PageSize = 1000

type DeviceListRequest struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
}

const (
	DEVICES_STATUS_ONLINE = iota + 1
	DEVICES_STATUS_FAULT
	DEVICES_STATUS_OFFLINE
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
	StationId          string `json:"stationId"`
	StationName        string `json:"stationName"`
	Status             int    `json:"status"`
	HasPV              bool   `json:"hasPV"`
	HasBattery         bool   `json:"hasBattery"`
	DeviceType         string `json:"deviceType"`
	ProductType        string `json:"productType"`
}

func (api *FoxessApi) GetDeviceList() ([]Device, error) {
	currentPage := 1
	total := 1
	devices := make([]Device, 0)
	for len(devices) < total {
		request := &DeviceListRequest{
			CurrentPage: currentPage,
			PageSize:    PageSize,
		}
		response := &DeviceListResponse{}
		if err := api.NewRequest("POST", "/op/v0/device/list", request, response); err != nil {
			return nil, err
		} else if err = IsError(response.ErrorNumber, ""); err != nil {
			return nil, err
		}

		devices = append(devices, response.Result.Devices...)
		total = response.Result.Total
		currentPage += 1
	}

	return devices, nil
}

func (d *Device) CurrentStatus() string {
	switch d.Status {
	case DEVICES_STATUS_ONLINE:
		return "Online"
	case DEVICES_STATUS_FAULT:
		return "Fault"
	case DEVICES_STATUS_OFFLINE:
		return "Offline"
	default:
		return fmt.Sprint("Unknown:", d.Status)
	}
}
