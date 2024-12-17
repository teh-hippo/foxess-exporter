package main

import (
	"fmt"

	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type DevicesCommand struct {
	FullOutput bool   `short:"f" long:"full" description:"Show all columns in the output" required:"false"`
	Format     string `short:"o" long:"output" description:"Output format" default:"table" choices:"table,json" required:"false"`
}

func init() {
	if _, err := parser.AddCommand("devices", "List devices", "Obtains all devices the provided key has access to", &DevicesCommand{}); err != nil {
		panic(err)
	}
}

func (x *DevicesCommand) Execute(_ []string) error {
	if x.Format == "json" && x.FullOutput {
		return fmt.Errorf("full output is not supported for JSON format")
	}

	devices, err := foxessApi.GetDeviceList()
	if err != nil {
		return fmt.Errorf("failed to retrieve device list: %w", err)
	}

	switch x.Format {
	case "table":
		x.OutputAsTable(devices)
		return nil
	case "json":
		return util.JsonToStdOut(devices)
	default:
		return fmt.Errorf("unsupported output format: %s", x.Format)
	}
}

func (x *DevicesCommand) OutputAsTable(devices []foxess.Device) {
	var tbl table.Table
	if x.FullOutput {
		tbl = table.New("Device Serial Number", "Module Serial Number", "Station ID", "Station Name", "Status", "Has PV", "Has Battery", "Device Type", "Product Type")
		for _, device := range devices {
			tbl.AddRow(device.DeviceSerialNumber, device.ModuleSerialNumber, device.StationId, device.StationName, device.CurrentStatus(), device.HasPV, device.HasBattery, device.DeviceType, device.ProductType)
		}
	} else {
		tbl = table.New("Device Serial Number", "Station Name", "Status", "Has PV", "Has Battery")
		for _, device := range devices {
			tbl.AddRow(device.DeviceSerialNumber, device.StationName, device.CurrentStatus(), device.HasPV, device.HasBattery)
		}
	}
	tbl.Print()
}
