package main

import (
	"encoding/json"
	"fmt"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
)

type DevicesCommand struct {
	FullOutput bool   `short:"f" long:"full" description:"Show all columns in the output" required:"false"`
	Format     string `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

type DeviceListRequest struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
}

const (
	DEVICES_STATUS_ONLINE  = 1
	DEVICES_STATUS_FAULT   = 2
	DEVICES_STATUS_OFFLINE = 3
)

const PageSize = 100

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

type DeviceListResponse struct {
	ErrorNumber int `json:"errno"`
	Result      struct {
		CurrentPage int      `json:"currentPage"`
		PageSize    int      `json:"pageSize"`
		Total       int      `json:"total"`
		Devices     []Device `json:"data"`
	}
}

var devicesCommand DevicesCommand

func init() {
	parser.AddCommand("devices", "List devices", "Obtains all devices the provided key has access to", &devicesCommand)
}

func (x *DevicesCommand) Execute(args []string) error {
	if x.Format == "json" && x.FullOutput {
		return fmt.Errorf("full output is not supported for JSON format")
	}

	devices, err := GetDeviceList()
	if err != nil {
		return err
	}

	switch x.Format {
	case "table":
		x.OutputAsTable(devices)
	case "json":
		return x.OutputAsJSON(devices)
	}
	return nil
}

func (x *DevicesCommand) OutputAsTable(devices []Device) {
	var tbl table.Table
	if x.FullOutput {
		tbl = table.New("Device Serial Number", "Module Serial Number", "Station ID", "Station Name", "Status", "Has PV", "Has Battery", "Device Type", "Product Type")
		for _, device := range devices {
			tbl.AddRow(device.DeviceSerialNumber, device.ModuleSerialNumber, device.StationId, device.StationName, IsOnline(device.Status), device.HasPV, device.HasBattery, device.DeviceType, device.ProductType)
		}
	} else {
		tbl = table.New("Device Serial Number", "Station Name", "Status", "Has PV", "Has Battery")
		for _, device := range devices {
			tbl.AddRow(device.DeviceSerialNumber, device.StationName, IsOnline(device.Status), device.HasPV, device.HasBattery)
		}
	}
	tbl.Print()
}

func (x *DevicesCommand) OutputAsJSON(devices []Device) error {
	for _, device := range devices {
		result, err := json.Marshal(device)
		if err != nil {
			return err
		}
		fmt.Println(string(result))
	}
	return nil
}

func IsOnline(status int) string {
	switch status {
	case DEVICES_STATUS_ONLINE:
		return "Online"
	case DEVICES_STATUS_FAULT:
		return "Fault"
	case DEVICES_STATUS_OFFLINE:
		return "Offline"
	default:
		return fmt.Sprint("Unknown:", status)
	}
}

func GetDeviceList() ([]Device, error) {
	currentPage := 1
	total := 1
	devices := make([]Device, 0)
	for len(devices) < total {
		request := &DeviceListRequest{
			CurrentPage: currentPage,
			PageSize:    PageSize,
		}
		response := &DeviceListResponse{}
		err := foxess.NewRequest(options.ApiKey, "POST", "/op/v0/device/list", request, response, options.Debug)

		if err != nil {
			return nil, err
		}

		if err = foxess.IsError(response.ErrorNumber, ""); err != nil {
			return nil, err
		}

		devices = append(devices, response.Result.Devices...)
		total = response.Result.Total
		currentPage += 1
	}

	return devices, nil
}
