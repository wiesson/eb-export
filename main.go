package main

import (
	"flag"
	"github.com/wiesson/eb-export/api"
	"github.com/wiesson/eb-export/config"
	"github.com/wiesson/eb-export/export"
	"time"
)

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func main() {
	start := Bod(time.Now().AddDate(0, 0, -2))
	to := start.AddDate(0, 0, 1)

	cmdToken := flag.String("cmdToken", "", "Access cmdToken")
	cmdFrom := flag.String("from", start.Format("2006-1-2"), "The lower date")
	cmdTo := flag.String("to", to.Format("2006-1-2"), "The upper date")
	cmdLogger := flag.String("cmdLogger", "", "Id of the data-cmdLogger")
	cmdAggregationLevel := flag.String("aggr", "", "Aggregation level")
	cmdFormat := flag.String("cmdFormat", "", "export json or csv")

	var cmdInputSensors config.Flags
	flag.Var(&cmdInputSensors, "sensor", "Id of the sensor")

	var cmdEnergyTypes config.Flags
	flag.Var(&cmdEnergyTypes, "type", "energy type")

	flag.Parse()

	apiConfig := config.New(
		*cmdToken,
		*cmdLogger,
		*cmdAggregationLevel,
		*cmdFrom,
		*cmdTo,
		*cmdFormat,
		cmdInputSensors,
		cmdEnergyTypes,
	)

	apiHandler := api.New(apiConfig)
	apiHandler.FetchLogger()
	sensors, samples, data := apiHandler.FetchSamples()

	writer := export.New(samples, sensors, data, apiConfig, *cmdFormat)
	writer.Write()
}
