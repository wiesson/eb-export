package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/wiesson/eb-export/api"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type argsWithSameKey []string

func (a *argsWithSameKey) String() string {
	return strings.Join(*a, ",")
}

func (a *argsWithSameKey) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func (a *argsWithSameKey) AsSlice() []string {
	slice := []string{}
	for _, item := range *a {
		slice = append(*a, item)
	}
	return slice
}

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func main() {
	start := Bod(time.Now().AddDate(0, 0, -2))
	from := start.AddDate(0, 0, 1)

	aggregationLevels := []string{"none", "minutes_1", "minutes_15", "hours_1", "days_1"}
	energyTypes := []string{"power", "energy"}

	cmdFrom := flag.String("from", start.Format("2006-1-2"), "The lower date")
	cmdTo := flag.String("to", from.Format("2006-1-2"), "The upper date")
	logger := flag.String("logger", "", "Id of the data-logger")
	tz := flag.String("tz", "UTC", "The identifier of the timezone, Europe/Berlin")
	aggregationLevel := flag.String("aggr", aggregationLevels[1], "Aggregation level")
	energyType := flag.String("type", energyTypes[0], "EnergyType")

	var sensors argsWithSameKey
	flag.Var(&sensors, "sensor", "Id of the data-logger")
	flag.Parse()

	if *aggregationLevel != aggregationLevels[2] {
		if inSlice(*aggregationLevel, aggregationLevels) == false {
			log.Fatal("Wrong aggregation level given. Valid levels are ", strings.Join(aggregationLevels, ", "))
			os.Exit(1)
		}
	}

	if *aggregationLevel != energyTypes[0] {
		if inSlice(*energyType, energyTypes) == false {
			log.Fatal("Wrong energyType given. Valid types are ", strings.Join(energyTypes, ", "))
			os.Exit(1)
		}
	}

	var loc, err = time.LoadLocation(*tz)
	if err != nil {
		fmt.Errorf("timezone could not be parsed: %v", err)
		os.Exit(1)
	}

	lower, _ := time.ParseInLocation("2006-1-2", *cmdFrom, loc)
	upper, _ := time.ParseInLocation("2006-1-2", *cmdTo, loc)

	log.Printf("You have entered %s %s %s and %d sensors\n", lower, upper, *logger, len(sensors))

	apiHandler := api.Config{
		DataLogger:       *logger,
		Sensors:          sensors.AsSlice(),
		TimeFrom:         lower.Unix(),
		TimeTo:           upper.Unix(),
		AggregationLevel: *aggregationLevel,
		Tz:               *loc,
		EnergyType:       *energyType,
	}

	data := &api.Data{}

	log.Println("Beginn Fetching data")

	nextUrl := apiHandler.GetRequestPath("")
	hasNext := true

	for hasNext {
		res, err := apiHandler.Get(nextUrl)
		if err != nil {
			log.Fatal(err)
		}

		logMessage, err := url.ParseQuery(nextUrl)
		log.Printf("Fetching from %s\n", logMessage.Get("page[offset]"))

		for _, value := range res.Data {
			data.AddItem(value, *energyType)
		}

		nextUrl = res.Links.NextURL
		if nextUrl == "" {
			hasNext = false
			break
		}
	}

	log.Println("Done")

	fileName := fmt.Sprintf("%s_%s_%s_%s_%s.csv", *cmdFrom, *cmdTo, *logger, *energyType, *aggregationLevel)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("error creating result.csv", err)
	}

	defer file.Close()

	w := csv.NewWriter(file)

	csvHeader := []string{"timestamp"}
	for _, sensorId := range sensors {
		csvHeader = append(csvHeader, sensorId)
	}

	if err := w.Write(csvHeader); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	for _, column := range *data {
		var csvRow []string
		csvRow = append(csvRow, column.DateTime.UTC().Format("2006-01-02 15:04:05"))
		for _, sensorID := range sensors {
			v := column.Readings[sensorID]
			csvRow = append(csvRow, v.String())
		}

		if err := w.Write(csvRow); err != nil {
			log.Fatalln("error writing row to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
	log.Printf("\nCreated file: %s\n", fileName)
}
