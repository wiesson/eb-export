package main

import (
	"flag"
	"github.com/wiesson/eb-export/api"
	"github.com/wiesson/eb-export/config"
	"github.com/wiesson/eb-export/export"
	"time"
)

var aggregationLevels = []string{"none", "minutes_1", "minutes_15", "hours_1", "days_1"}
var energyTypes = []string{"power", "energy"}

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func main() {
	start := Bod(time.Now().AddDate(0, 0, -2))
	to := start.AddDate(0, 0, 1)

	cmdFrom := flag.String("from", start.Format("2006-1-2"), "The lower date")
	cmdTo := flag.String("to", to.Format("2006-1-2"), "The upper date")
	logger := flag.String("logger", "", "Id of the data-logger")
	tz := flag.String("tz", "UTC", "The identifier of the timezone, Europe/Berlin")
	aggregationLevel := flag.String("aggr", aggregationLevels[1], "Aggregation level")
	energyType := flag.String("type", energyTypes[0], "EnergyType")

	var sensors config.Flags
	flag.Var(&sensors, "sensor", "Id of the data-logger")
	flag.Parse()

	apiConfig := config.New(
		*logger,
		*energyType,
		*aggregationLevel,
		*tz,
		*cmdFrom,
		*cmdTo,
		sensors,
		aggregationLevels,
		energyTypes,
	)

	apiHandler := api.New(apiConfig)
	data := apiHandler.Fetch()

	writer := export.New(data, apiConfig)
	writer.Write()
}
