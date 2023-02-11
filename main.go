package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listenAddress := flag.String("listen.address", ":28272", "Listening address for metrics")
	metricsPath := flag.String("metrics.path", "/metrics", "Path under which to expose metrics")
	configFile := flag.String("config.file", "config.yaml", "Configuration file describing Nagios checks")
	configExample := flag.Bool("config.example", false, "Show example configuration file")
	logLevel := flag.Int("log.level", 1, "Log level: 0 (ERROR), 1 (INFO), 2 (DEBUG)")

	flag.Parse()

	if *configExample {
		ShowConfigExample()
		os.Exit(0)
	}

	InitLogger(*logLevel)

	config, err := ReadConfig(*configFile)
	if err != nil {
		logger.ERROR.Fatalln(err)
	}

	if len(config.Commands) == 0 {
		logger.ERROR.Fatalln("no commands defined")
	}

	exits := make(chan bool)

	for _, command := range config.Commands {
		go checkLooper(command, exits)
	}

	go listenForExits(len(config.Commands), exits)

	logger.INFO.Println("listening on", *listenAddress)

	http.Handle(*metricsPath, HTTPWithLogger(promhttp.Handler()))

	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		logger.ERROR.Fatalln(err)
	}
}
