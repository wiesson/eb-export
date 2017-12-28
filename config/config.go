package config

import (
	"log"
	"os"
	"strings"
	"time"
)

var aggregationLevels = []string{"none", "minutes_1", "minutes_15", "hours_1", "days_1"}
var exportFileFormats = []string{"json", "csv"}
var defaultEnergyType = []string{"power"}

// Config contains the configuration
type Config struct {
	AccessToken      string
	DataLogger       string
	EnergyTypes      []string
	InputSensors     []string
	TimeFrom         time.Time
	TimeTo           time.Time
	AggregationLevel string
	Format           string
}

// New returns a new instance of Config
func New(cmdToken, cmdDataLogger, cmdAggregation, cmdFrom, cmdTo, cmdFormat string, cmdInputSensors, cmdEnergyTypes []string) Config {
	if cmdToken == "" {
		log.Fatal("No access token given.")
		os.Exit(1)
	}

	if inSlice(cmdAggregation, aggregationLevels) == false {
		cmdAggregation = aggregationLevels[1]
	}

	if cmdFormat == "" {
		cmdFormat = exportFileFormats[0]
	}

	if len(cmdEnergyTypes) == 0 {
		cmdEnergyTypes = defaultEnergyType
	}

	cmdTimeFrom, err := time.Parse("2006-1-2T15:04:05", completeDate(cmdFrom))
	if err != nil {
		log.Fatal("could not parse from date")
		os.Exit(1)
	}

	cmdTimeTo, err := time.Parse("2006-1-2T15:04:05", completeDate(cmdTo))
	if err != nil {
		log.Fatal("could not parse to date")
		os.Exit(1)
	}

	if cmdTimeTo.Before(cmdTimeFrom) {
		log.Fatal("from date is before to date")
		os.Exit(1)
	}

	if len(cmdInputSensors) > 0 {
		log.Printf("You have entered %s %s %s and %d sensors\n", cmdTimeFrom, cmdTimeTo, cmdDataLogger, len(cmdInputSensors))
	} else {
		log.Printf("You have entered %s %s %s and all sensors\n", cmdTimeFrom, cmdTimeTo, cmdDataLogger)
	}

	return Config{
		AccessToken:      cmdToken,
		Format:           cmdFormat,
		DataLogger:       cmdDataLogger,
		EnergyTypes:      cmdEnergyTypes,
		InputSensors:     cmdInputSensors,
		TimeFrom:         cmdTimeFrom,
		TimeTo:           cmdTimeTo,
		AggregationLevel: cmdAggregation,
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

func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func completeDate(date string) string {
	length := len(date)
	if length == 10 {
		return date + "T00:00:00"
	}
	return date
}
