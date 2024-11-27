package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServeCommand struct {
	Port int `short:"p" long:"port" description:"Port to listen on" default:"2112"`
}

var serveCommand ServeCommand

func init() {
	parser.AddCommand("serve", "Start the exporter", "Start the exporter", &serveCommand)
}

func (x *ServeCommand) Execute(args []string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":2112", nil)
}
