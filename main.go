package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/teh-hippo/foxess-exporter/foxess"
)

func main() {
	apiKey := flag.String("api-key", "", "FoxESS API Key")
	inverter := flag.String("inverter", "", "Inverter to target")

	flag.Parse()
	if len(*apiKey) == 0 {
		fmt.Fprint(os.Stderr, "Missing an API Key\n\n")
		flag.Usage()
		os.Exit(1)
	}

	err := foxess.GetHistory(*apiKey, *inverter)
	if err != nil {
		log.Fatalf("Error: %s\n", err)
		os.Exit(1)
	}
}
