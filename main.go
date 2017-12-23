package main

import (
	"flag"
	"github.com/wiesson/eb-export/api"
	"github.com/wiesson/eb-export/config"
	"github.com/wiesson/eb-export/export"
	"time"
)

var aggregationLevels = []string{"none", "minutes_1", "minutes_15", "hours_1", "days_1"}
var exportFileFormats = []string{"json", "csv"}

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func main() {
	start := Bod(time.Now().AddDate(0, 0, -2))
	to := start.AddDate(0, 0, 1)

	token := flag.String("token", "", "Access token")
	cmdFrom := flag.String("from", start.Format("2006-1-2"), "The lower date")
	cmdTo := flag.String("to", to.Format("2006-1-2"), "The upper date")
	logger := flag.String("logger", "", "Id of the data-logger")
	aggregationLevel := flag.String("aggr", aggregationLevels[1], "Aggregation level")
	format := flag.String("format", exportFileFormats[0], "export json or csv")

	var inputSensors config.Flags
	flag.Var(&inputSensors , "sensor", "Id of the sensor")

	var energyTypes config.Flags
	flag.Var(&energyTypes, "type", "energy type")

	flag.Parse()

	apiConfig := config.New(
		*token,
		*logger,
		*aggregationLevel,
		*cmdFrom,
		*cmdTo,
		*format,
		inputSensors,
		energyTypes,
		aggregationLevels,
	)

	apiHandler := api.New(apiConfig)
	apiHandler.FetchLogger()
	sensors, samples, data := apiHandler.FetchSamples()

	writer := export.New(samples, sensors, data, apiConfig, *format)
	writer.Write()
}
