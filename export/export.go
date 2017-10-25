package export

import (
	"encoding/csv"
	"fmt"
	"github.com/wiesson/eb-export/api"
	"github.com/wiesson/eb-export/config"
	"log"
	"os"
)

type Export struct {
	sensors  []string
	fileName string
	data     api.Data
}

func New(data api.Data, apiConfig config.Config) Export {
	name := GetFileName(apiConfig)

	return Export{
		sensors:  apiConfig.Sensors,
		fileName: name,
		data:     data,
	}
}

func (e *Export) Write() {
	file, err := os.Create(e.fileName)
	if err != nil {
		log.Fatalln("error creating result.csv", err)
	}

	defer file.Close()

	w := csv.NewWriter(file)

	csvHeader := []string{"timestamp"}
	for _, sensorId := range e.sensors {
		csvHeader = append(csvHeader, sensorId)
	}

	if err := w.Write(csvHeader); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	for _, column := range e.data {
		var csvRow []string
		csvRow = append(csvRow, column.DateTime.UTC().Format("2006-01-02 15:04:05"))
		for _, sensorID := range e.sensors {
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
	log.Printf("Created file: %s\n", e.fileName)
}

func GetFileName(apiConfig config.Config) string {
	return fmt.Sprintf("%d_%d_%s_%s_%s.csv", apiConfig.TimeTo, apiConfig.TimeTo, apiConfig.DataLogger, apiConfig.EnergyType, apiConfig.AggregationLevel)
}
