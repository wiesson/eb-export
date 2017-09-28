package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"github.com/wiesson/eb-export/samples"
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

func main() {
	start := Bod(time.Now().AddDate(0, 0, -2))
	from := start.AddDate(0, 0, 1)

	cmdFrom := flag.String("from", start.Format("2006-1-2"), "The lower date")
	cmdTo := flag.String("to", from.Format("2006-1-2"), "The upper date")
	logger := flag.String("logger", "", "Id of the data-logger")
	tz := flag.String("tz", "UTC", "The identifier of the timezone, Europe/Berlin")
	aggregationLevel := flag.String("aggr", "minutes_1", "Aggregation level")
	energyType := flag.String("type", "power", "EnergyType")

	var sensors argsWithSameKey
	flag.Var(&sensors, "sensor", "Id of the data-logger")
	flag.Parse()

	var loc, err = time.LoadLocation(*tz)
	if err != nil {
		fmt.Errorf("timezone could not be parsed: %v", err)
		os.Exit(1)
	}

	lower, _ := time.ParseInLocation("2006-1-2", *cmdFrom, loc)
	upper, _ := time.ParseInLocation("2006-1-2", *cmdTo, loc)

	fmt.Printf("You have entered %s %s %s and %d sensors\n", lower, upper, *logger, len(sensors))

	api := samples.API{
		DataLogger:       *logger,
		Sensors:          sensors.AsSlice(),
		TimeFrom:         lower.Unix(),
		TimeTo:           upper.Unix(),
		AggregationLevel: *aggregationLevel,
		Tz:               *loc,
		EnergyType:       *energyType,
	}

	data := &samples.Data{}

	fmt.Print("Fetching data")

	nextUrl := api.GetRequestPath("")
	hasNext := true

	for hasNext {
		res, err := api.Get(nextUrl)
		if err != nil {
			log.Fatal(err)
		}

		for _, value := range res.Data {
			data.AddItem(value, *energyType)
		}

		nextUrl = res.Links.NextURL
		if nextUrl == "" {
			hasNext = false
			break
		}

		fmt.Print(".")
	}

	fmt.Print(" Done")

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
	fmt.Printf("\nCreated file: %s\n", fileName)
}
