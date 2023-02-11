package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

type PerfData struct {
	Units string
	Value float64
	Min   float64
	Max   float64
}

type Output struct {
	Message  string
	Status   int
	PerfData map[string]PerfData
	Duration time.Duration
}

func listenForExits(count int, exits chan bool) {
	for range exits {
		count -= 1
		if count <= 0 {
			logger.INFO.Fatalln("all commands stopped")
		}
	}
}

func checkLooper(command *ConfigCommand, exits chan bool) {
	defer func() {
		exits <- true
	}()

	ticker := time.NewTicker(command.CheckInterval)
	defer ticker.Stop()

	err := checkRunner(command)
	if err != nil {
		return
	}

	for range ticker.C {
		err := checkRunner(command)
		if err != nil {
			break
		}
	}
}

func checkRunner(command *ConfigCommand) error {
	out := Output{
		PerfData: make(map[string]PerfData),
	}

	ctx, cancel := context.WithTimeout(context.Background(), command.Timeout)
	defer cancel()

	logger.DEBUG.Println("running command", strconv.Quote(command.Name))

	cmd := exec.CommandContext(ctx, command.Command[0], command.Command[1:]...)
	start := time.Now()
	data, err := cmd.Output()
	out.Duration = time.Since(start)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.ERROR.Println("command timed out", strconv.Quote(command.Name))

			return nil
		}

		logger.ERROR.Println("command", strconv.Quote(command.Name), "error:", err)

		return fmt.Errorf("running command: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 || i == len(lines)-2 { // first or second to last
			tmp := strings.SplitN(line, "|", 2)
			line = tmp[0]

			if len(tmp) > 1 {
				maps.Copy(out.PerfData, splitPerfData(tmp[1]))
			}
		}

		out.Message += line + "\n"
	}

	out.Message = strings.TrimSpace(out.Message)
	out.Status = cmd.ProcessState.ExitCode()

	logger.DEBUG.Printf(
		"finished command %s in %fs with status %d\n",
		strconv.Quote(command.Name), out.Duration.Seconds(), out.Status,
	)
	RecordMetric(command, out)

	return nil
}

func splitQuoted(line string, sep rune) []string {
	var out []string

	inQuote := false
	inSep := false
	last := 0

	for idx, char := range line {
		if char != sep {
			inSep = false
		}

		if char == '\'' {
			inQuote = !inQuote

			continue
		}

		if inQuote {
			continue
		}

		if char == sep {
			if !inSep {
				item := line[last:idx]
				if len(item) > 0 {
					out = append(out, item)
				}
			}

			last = idx + 1
			inSep = true
		}
	}

	item := line[last:]
	if len(item) > 0 {
		out = append(out, item)
	}

	return out
}

func splitPerfData(line string) map[string]PerfData {
	out := make(map[string]PerfData)
	items := splitQuoted(line, ' ')

	for _, item := range items {
		tmp := splitQuoted(item, '=')

		if len(tmp) < 2 {
			// just label, that's weird
			continue
		}

		label := strings.Trim(tmp[0], "'")

		tmp = strings.Split(tmp[1], ";")

		max := math.NaN()
		min := math.NaN()
		value := math.NaN()
		units := ""

		var err error

		switch {
		case len(tmp) >= 5:
			max, err = strconv.ParseFloat(tmp[4], 64)
			if err != nil {
				max = math.NaN()
			}

			fallthrough

		case len(tmp) >= 4:
			min, err = strconv.ParseFloat(tmp[3], 64)
			if err != nil {
				min = math.NaN()
			}

			fallthrough

		case len(tmp) >= 1:
			valueStr := tmp[0]
			units = strings.TrimLeft(valueStr, "0123456789.")
			valueStr = valueStr[:len(valueStr)-len(units)]

			value, err = strconv.ParseFloat(valueStr, 64)
			if err != nil {
				value = math.NaN()
			}
		}

		out[label] = PerfData{
			Units: units,
			Value: value,
			Min:   min,
			Max:   max,
		}
	}

	return out
}
