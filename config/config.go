package config

import (
	"log"
	"os"
	"strings"
	"time"
)

var exportFileFormats = []string{"json", "csv"}
var defaultEnergyType = []string{"power"}

type Aggregation struct {
	Level    string
	Interval time.Duration
}

var aggregationTypes = map[string]Aggregation{
	"none":      {"none", time.Second * 2},
	"minutes_1": {"minutes_1", time.Minute},
	"hours_1":   {"hours_1", time.Hour},
	"days_1":    {"days_1", time.Hour * 24},
}

// Config contains the configuration
type Config struct {
	AccessToken  string
	DataLogger   string
	EnergyTypes  []string
	InputSensors []string
	TimeFrom     time.Time
	TimeTo       time.Time
	Format       string
	Aggregation
}

// New returns a new instance of Config
func New(cmdDataLogger, cmdAggregation, cmdFrom, cmdTo, cmdFormat string, cmdInputSensors, cmdEnergyTypes []string) Config {
	cmdToken := os.Getenv("EB_ACCESS_TOKEN")
	if cmdToken == "" {
		log.Fatal("No access token given.")
		os.Exit(1)
	}

	aggregation, ok := aggregationTypes[cmdAggregation]
	if ok != true {
		aggregation = aggregationTypes["minutes_1"]
	}

	if cmdFormat == "" {
		cmdFormat = exportFileFormats[0]
	}

	if len(cmdEnergyTypes) == 0 {
		cmdEnergyTypes = defaultEnergyType
	}

	cmdTimeFrom, err := time.Parse(time.RFC3339, completeDate(cmdFrom))
	if err != nil {
		log.Fatal("could not parse from date")
		os.Exit(1)
	}

	cmdTimeTo, err := time.Parse(time.RFC3339, completeDate(cmdTo))
	if err != nil {
		log.Fatal("could not parse to date")
		os.Exit(1)
	}
	if cmdTimeTo.Before(cmdTimeFrom) {
		log.Fatal("from date is before to date")
		os.Exit(1)
	}

	if len(cmdInputSensors) > 0 {
		log.Printf("From: %s, To: %s, %s and %d sensors\n", cmdTimeFrom, cmdTimeTo, cmdDataLogger, len(cmdInputSensors))
	} else {
		log.Printf("From: %s, To: %s, %s and all sensors\n", cmdTimeFrom, cmdTimeTo, cmdDataLogger)
	}

	return Config{
		AccessToken:  cmdToken,
		Format:       cmdFormat,
		DataLogger:   cmdDataLogger,
		EnergyTypes:  cmdEnergyTypes,
		InputSensors: cmdInputSensors,
		TimeFrom:     cmdTimeFrom,
		TimeTo:       cmdTimeTo,
		Aggregation:  aggregation,
	}
}

// Flags is an helper for flags with same argument
type Flags []string

func (f *Flags) String() string {
	return strings.Join(*f, ",")
}

func (f *Flags) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f *Flags) Slice() []string {
	var slice []string
	for _, item := range *f {
		slice = append(*f, item)
	}
	return slice
}

func completeDate(date string) string {
	length := len(date)
	if length == 10 {
		return date + "T00:00:00+00:00"
	}
	return date
}

// Bod returns the beginning of a day
func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func DefaultLowerTime() time.Time {
	return Bod(time.Now().AddDate(0, 0, -2))
}

func DefaultUpperTime() time.Time {
	return DefaultLowerTime().AddDate(0, 0, 1)
}
