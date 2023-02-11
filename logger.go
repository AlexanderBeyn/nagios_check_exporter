package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

type Logger struct {
	ERROR *log.Logger // 0
	INFO  *log.Logger // 1
	DEBUG *log.Logger // 2
}

var logger = Logger{
	DEBUG: log.New(io.Discard, "nagios_check_exporter: DEBUG: ", log.Ldate|log.Ltime|log.Lmsgprefix),
	INFO:  log.New(io.Discard, "nagios_check_exporter: INFO: ", log.Ldate|log.Ltime|log.Lmsgprefix),
	ERROR: log.New(io.Discard, "nagios_check_exporter: ERROR: ", log.Ldate|log.Ltime|log.Lmsgprefix),
}

func InitLogger(level int) {
	switch {
	case level >= 2:
		logger.DEBUG.SetOutput(os.Stdout)

		fallthrough

	case level == 1:
		logger.INFO.SetOutput(os.Stdout)

		fallthrough

	case level == 0:
		logger.ERROR.SetOutput(os.Stdout)
	}
}

func HTTPWithLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		logger.DEBUG.Printf("HTTP %s from %s: %s\n", r.Method, r.RemoteAddr, r.URL.String())
	})
}
