package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/cobra"
	vegeta "github.com/tsenart/vegeta/lib"
	"github.com/valencenet/majin/internal/generate"
)

// headers is the http.Header used in each target request
// it is defined here to implement the flag.Value interface
// in order to support multiple identical flags for request header
// specification
type headers struct{ http.Header }

func (h *headers) Type() string {
	return "headers"
}

func (h *headers) String() string {
	buf := &bytes.Buffer{}
	if err := h.Write(buf); err != nil {
		return ""
	}
	return buf.String()
}

// Set implements the flag.Value interface for a map of HTTP Headers.
func (h *headers) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("header '%s' has a wrong format", value)
	}
	key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if key == "" || val == "" {
		return fmt.Errorf("header '%s' has a wrong format", value)
	}
	// Add key/value directly to the http.Header (map[string][]string).
	// http.Header.Add() canonicalizes keys but vegeta is used
	// to test systems that require case-sensitive headers.
	h.Header = make(http.Header)
	h.Header[key] = append(h.Header[key], val)
	return nil
}

var (
	simulateCmd = &cobra.Command{
		Use:   "attack",
		Short: "Generate a workload simulation and attack according to the simulation parameters given",
		RunE:  simulateCommand,
	}

	simulateFlags struct {
		Random bool `long:"random" env:"RANDOM" description:"Generate random workloads? Use the other parameters as max random values."`

		Load           float64 `long:"base-load" env:"BASE_LOAD" description:"The base queries per second to use for attack" default:"1"`
		Period         float64 `long:"period" env:"PERIOD" description:"The duration of each period in the simulated timeseries" default:"60"`
		Duration       float64 `long:"duration" env:"DURATION" description:"The duration of the total attack" default:"0"`
		Trend          float64 `long:"trend" env:"TREND" description:"The trend degree with which the timeseries increases (or decreases with negative)" default:"0"`
		Noise          float64 `long:"noise" env:"NOISE" description:"The noise factor for the timeseries" default:"0"`
		BurstFrequency float64 `long:"burst-frequency" env:"BURST_FREQUENCY" description:"Frequency of  bursts - 1 in frequency" default:"0"`
		BurstIndex     float64 `long:"burst-index" env:"BURST_INDEX" description:" bursts multiplier" default:"1"`
		BurstDuration  float64 `long:"burst-duration" env:"BURST_DURATION" description:"The duration of the burst" default:"30"`
		Target         string  `long:"target" env:"TARGET" description:"The target to point the workload against"`

		Method string  `long:"method" env:"METHOD" description:"The http method to use" default:"GET"`
		Header headers `long:"header" env:"HEADER" description:"The http headers"`
		Body   string  `long:"body" env:"BODY" description:"The http body to send"`
	}
)

func simulateCommand(cmd *cobra.Command, args []string) error {
	if _, err := flags.ParseArgs(&simulateFlags, append(args, os.Args...)); err != nil {
		log.Printf("Could not parse flags: %s", err)
		return err
	}

	log.Printf("Launching attack on %s", simulateFlags.Target)
	target := vegeta.Target{
		Method: simulateFlags.Method,
		URL:    simulateFlags.Target,
		Header: simulateFlags.Header.Header,
	}
	if simulateFlags.Body != "" {
		target.Body = []byte(simulateFlags.Body)
	}
	targeter := vegeta.NewStaticTargeter(target)
	attacker := vegeta.NewAttacker()

	if !simulateFlags.Random {
		workload := generate.Workload{
			Load:           simulateFlags.Load,
			Trend:          simulateFlags.Trend,
			Period:         simulateFlags.Period,
			Noise:          simulateFlags.Noise,
			BurstIndex:     simulateFlags.BurstIndex,
			BurstFrequency: simulateFlags.BurstFrequency,
			BurstDuration:  simulateFlags.BurstDuration,
			Duration:       simulateFlags.Duration,
		}
		workload.Simulation(attacker, targeter)
	} else {
		for {
			rand.Seed(time.Now().UTC().UnixNano())
			workload := generate.Workload{
				Load:           randInt64(simulateFlags.Load),
				Trend:          randInt64(simulateFlags.Trend),
				Period:         randInt64(simulateFlags.Period),
				Noise:          randInt64(simulateFlags.Noise),
				BurstIndex:     randInt64(simulateFlags.BurstIndex),
				BurstFrequency: randInt64(simulateFlags.BurstFrequency),
				BurstDuration:  randInt64(simulateFlags.BurstDuration),
				Duration:       randInt64(simulateFlags.Duration),
			}
			log.Printf("Performing the following attack with parameters: %+v", workload)
			workload.Simulation(attacker, targeter)
		}
	}
	return nil
}

func randInt64(n float64) float64 {
	if n != 0.0 {
		return float64(rand.Int63n(int64(n)))
	}
	return n
}

func init() {
	simulateCmd.Flags().BoolVar(&simulateFlags.Random, "random", false, "Generate random workloads? Use the other parameters as max random values.")

	simulateCmd.Flags().StringVar(&simulateFlags.Method, "method", "GET", "The http method to use")
	simulateCmd.Flags().Var(&simulateFlags.Header, "header", "The HTTP headers to use")
	simulateCmd.Flags().StringVar(&simulateFlags.Body, "body", "", "The body to send in request")

	simulateCmd.Flags().StringVar(&simulateFlags.Target, "target", "http://localhost:8080", "The target url to attack")
	simulateCmd.Flags().Float64Var(&simulateFlags.Period, "period", 60, "how long each period is - as int of seconds")
	simulateCmd.Flags().Float64Var(&simulateFlags.Duration, "duration", 0, "Duration to run the attack for. 0 means run the attack forever")
	simulateCmd.Flags().Float64Var(&simulateFlags.Trend, "trend", 0, "The degree of trend to attack with")
	simulateCmd.Flags().Float64Var(&simulateFlags.Noise, "noise", 0, "how much noise is in the workload")
	simulateCmd.Flags().Float64Var(&simulateFlags.BurstFrequency, "burst-frequency", 0, "how frequent  bursts should be")
	simulateCmd.Flags().Float64Var(&simulateFlags.BurstDuration, "burst-duration", 30, "how long  bursts should be")
	simulateCmd.Flags().Float64Var(&simulateFlags.BurstIndex, "burst-index", 0, " burst multiplier")
	simulateCmd.Flags().Float64Var(&simulateFlags.Load, "base-load", 1, "The base queries per second to use for attack")

	rootCmd.AddCommand(simulateCmd)
}
