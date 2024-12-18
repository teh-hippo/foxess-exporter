package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/rodaine/table"
	"github.com/teh-hippo/foxess-exporter/foxess"
	"github.com/teh-hippo/foxess-exporter/util"
)

type DevicesCommand struct {
	FullOutput bool   `short:"f" long:"full"   description:"Show all columns in the output"`
	Format     string `short:"o" long:"output" description:"Output format"                  default:"table" choices:"table,json"`
	config     *foxess.Config
}

func (x *DevicesCommand) Register(parser *flags.Parser, config *foxess.Config) {
	if _, err := parser.AddCommand("devices", "List devices", "Obtains all devices the provided key has access to", x); err != nil {
		panic(err)
	}

	x.config = config
}

func (x *DevicesCommand) Execute(_ []string) error {
	if x.Format == FormatJSON && x.FullOutput {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, "full output is not supported for JSON format")
	}

	devices, err := x.config.GetDeviceList()
	if err != nil {
		return fmt.Errorf("failed to retrieve device list: %w", err)
	}

	switch x.Format {
	case FormatTable:
		x.OutputAsTable(devices)

		return nil
	case FormatJSON:
		if err := util.JSONToStdOut(devices); err != nil {
			return fmt.Errorf("failed to output device list: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedFormat, x.Format)
	}
}

func (x *DevicesCommand) OutputAsTable(devices []foxess.Device) {
	var tbl table.Table
	if x.FullOutput {
		tbl = table.New("Device Serial Number", "Module Serial Number", "Station ID", "Station Name", "Status", "Has PV", "Has Battery", "Device Type", "Product Type")
		for _, device := range devices {
			tbl.AddRow(device.DeviceSerialNumber, device.ModuleSerialNumber, device.StationID, device.StationName, device.CurrentStatus(), device.HasPV, device.HasBattery,
				device.DeviceType, device.ProductType)
		}
	} else {
		tbl = table.New("Device Serial Number", "Station Name", "Status", "Has PV", "Has Battery")
		for _, device := range devices {
			tbl.AddRow(device.DeviceSerialNumber, device.StationName, device.CurrentStatus(), device.HasPV, device.HasBattery)
		}
	}

	tbl.Print()
}
