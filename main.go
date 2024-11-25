package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/teh-hippo/foxess-exporter/foxess"
)

func main() {
	apiKey := flag.String("api-key", "", "FoxESS API Key")
	inverter := flag.String("inverter", "", "Inverter to target")

	flag.Parse()
	if len(*apiKey) == 0 || len(*inverter) == 0 {
		fmt.Fprint(os.Stderr, "Missing an API Key and Inverter\n\n")
		flag.Usage()
		os.Exit(1)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)
	// .Date().Add(-24 * time.Hour)
	data := &foxess.Foxess{ApiKey: *apiKey, Inverter: *inverter}
	err := data.GetHistory(yesterday, today)
	if err != nil {
		log.Fatalf("Error: %s\n", err)
		os.Exit(1)
	}
}
