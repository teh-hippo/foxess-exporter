package main

import (
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"time"

	"github.com/teh-hippo/foxess-exporter/foxess"
)

func main() {
	apiKey := flag.String("api-key", "", "FoxESS API Key")
	inverter := flag.String("inverter", "", "Inverter to target")
	debug := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	if len(*apiKey) == 0 || len(*inverter) == 0 {
		fmt.Fprint(os.Stderr, "Missing an API Key and Inverter\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if flag.Args() == nil || len(flag.Args()) != 1 {
		fmt.Fprint(os.Stderr, "Missing command: <variables/history>\n\n")
		flag.Usage()
		os.Exit(1)
	}

	data := &foxess.Foxess{ApiKey: *apiKey, Inverter: *inverter, Debug: *debug}

	var err error
	switch flag.Args()[0] {
	case "variables":
		err = GetVariables(data)
	case "history":
		err = GetHistory(data)
	}
	if err != nil {
		log.Fatalf("Error: %s\n", err)
		os.Exit(1)
	}
}

func GetHistory(data *foxess.Foxess) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)
	return data.GetHistory(yesterday, today)
}

func GetVariables(data *foxess.Foxess) error {
	variables, err := data.GetAvailableVariables()
	if err != nil {
		return err
	}
	fmt.Printf("Variable Name\tDescription\tUnit\tGridTiedInverter\tEnergyStorageInverter\n")
	for _, variable := range variables.Result {
		for key := range maps.Keys(variable) {
			item := variable[key]
			fmt.Printf("%s\t%s\t%s\t%t\t%t\n", key, item.Name["en"], item.Unit, item.GridTiedInverter, item.EnergyStorageInverter)
		}
	}
	return nil
}
