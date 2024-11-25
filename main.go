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
	apiKey := flag.String("apiKey", "", "FoxESS API Key")
	path := flag.String("path", "", "Path")
	time := time.Now().Unix()
	testPath := "/op/v0/device/list"
	flag.Parse()
	if len(*apiKey) == 0 || len(*path) == 0 {
		fmt.Fprint(os.Stderr, "Missing an API Key and Path\n\n")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Key: %s, Path: %s, Time: %d\n", *apiKey, *path, time)
	log.Println(foxess.CalculateSignature(testPath, *apiKey, time))
}
