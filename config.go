package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type ConfigCommand struct {
	Name          string            `yaml:"name"`
	Command       []string          `yaml:"command"`
	Labels        map[string]string `yaml:"labels"`
	CheckInterval time.Duration     `yaml:"check_interval"`
	Timeout       time.Duration     `yaml:"timeout"`
}

type Config struct {
	Defaults ConfigCommand    `yaml:"defaults"`
	Commands []*ConfigCommand `yaml:"commands"`
}

func ReadConfig(file string) (*Config, error) {
	logger.DEBUG.Println("reading config file from", file)

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	out := &Config{}

	err = yaml.Unmarshal(data, out)
	if err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	for _, command := range out.Commands {
		command.CheckInterval = firstNonZero(command.CheckInterval, out.Defaults.CheckInterval, 5*time.Minute)
		command.Timeout = firstNonZero(command.Timeout, out.Defaults.Timeout, 30*time.Second)
		command.Labels = mergeMaps(out.Defaults.Labels, command.Labels)
		logger.DEBUG.Printf("command %s configured: %s\n", strconv.Quote(command.Name), strings.Join(command.Command, " "))
	}

	return out, nil
}

func firstNonZero[T comparable](args ...T) T {
	for _, arg := range args {
		if arg != reflect.Zero(reflect.TypeOf(arg)).Interface() {
			return arg
		}
	}

	return args[len(args)-1]
}

func mergeMaps[M ~map[K]V, K comparable, V any](mm ...M) M {
	out := make(M)
	for _, m := range mm {
		if m == nil {
			continue
		}
		maps.Copy(out, m)
	}

	return out
}

const exampleConfig = `# example configuration file
defaults:
  check_interval: 5m
  timeout: 60s
  labels:
    common_label: some_value

commands:
  - name: http
    command: [/usr/lib/nagios/plugins/check_http, -H, www.example.com]
    timeout: 10s
    labels:
      host: www.example.com
`

func ShowConfigExample() {
	fmt.Print(exampleConfig)
}
