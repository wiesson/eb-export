package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Config contains the configuration
type Config struct {
	AccessToken      string
	DataLogger       string
	Sensors          []string
	TimeFrom         int64
	TimeTo           int64
	AggregationLevel string
	Tz               time.Location
	EnergyType       string
}

// New returns a new instance of Config
func New(accessToken, DataLogger, energyType, aggregationLevel, timezone, cmdFrom, cmdTo string, sensors, aggregationLevels, energyTypes []string) Config {
	if accessToken == "" {
		log.Fatal("No access token given.")
		os.Exit(1)
	}

	if aggregationLevel != aggregationLevels[2] && inSlice(aggregationLevel, aggregationLevels) == false {
		log.Fatal("Wrong aggregation level given. Valid levels are ", strings.Join(aggregationLevels, ", "))
		os.Exit(1)
	}

	if aggregationLevel != energyTypes[0] && inSlice(energyType, energyTypes) == false {
		log.Fatal("Wrong energyType given. Valid types are ", strings.Join(energyTypes, ", "))
		os.Exit(1)
	}

	var loc, err = time.LoadLocation(timezone)
	if err != nil {
		fmt.Errorf("timezone could not be parsed: %v", err)
		os.Exit(1)
	}

	lower, _ := time.ParseInLocation("2006-1-2", cmdFrom, loc)
	upper, _ := time.ParseInLocation("2006-1-2", cmdTo, loc)

	log.Printf("You have entered %s %s %s and %d sensors\n", lower, upper, DataLogger, len(sensors))

	return Config{
		AccessToken:      accessToken,
		DataLogger:       DataLogger,
		Sensors:          sensors,
		TimeFrom:         lower.Unix(),
		TimeTo:           upper.Unix(),
		AggregationLevel: aggregationLevel,
		Tz:               *loc,
		EnergyType:       energyType,
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
	slice := []string{}
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
