package export

import (
	"fmt"
	"github.com/wiesson/eb-export/api"
	"github.com/wiesson/eb-export/config"
	"log"
	"os"
	"encoding/json"
	"strings"
)

type Export struct {
	sensors  []string
	fileName string
	fileType string
	// data     api.Data
	samples  api.Samples
}

func New(samples api.Samples, apiConfig config.Config, fileType string) Export {
	name := getFileName(apiConfig, fileType)

	return Export{
		sensors:  apiConfig.Sensors,
		fileName: name,
		samples:  samples,
		fileType: fileType,
	}
}

func (e *Export) Write() {
	if e.fileType == "json" {
		e.JSON()
	}

	/* if e.fileType == "csv" {
		e.CSV()
	} */
}

func (e *Export) JSON() {
	file, err := os.Create(e.fileName)
	if err != nil {
		log.Fatalln("error creating export.json", err)
	}

	defer file.Close()

	samplesJson, _ := json.Marshal(e.samples)
	file.Write(samplesJson)
	log.Printf("Created file: %s\n", e.fileName)
}
/*
func (e *Export) CSV() {
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

			if v == nil {
				csvRow = append(csvRow, "")
			} else {
				csvRow = append(csvRow, strconv.FormatFloat(*v, 'f', 8, 64))
			}

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
} */

func getFileName(apiConfig config.Config, fileType string) string {
	energyTypes := strings.Join(apiConfig.EnergyTypes, "-")
	return fmt.Sprintf("%d_%d_%s_%s_%s.%s", apiConfig.TimeFrom.Unix(), apiConfig.TimeTo.Unix(), apiConfig.DataLogger, energyTypes, apiConfig.AggregationLevel, fileType)
}
