package export

import (
	"fmt"
	"github.com/wiesson/eb-export/api"
	"github.com/wiesson/eb-export/config"
	"log"
	"os"
	"encoding/json"
	"strings"
	"encoding/csv"
	"strconv"
)

type Export struct {
	energyTypes     []string
	selectedSensors []api.Sensor
	samples         api.Samples
	data            api.Data
	fileName        string
	fileType        string
}

func New(samples api.Samples, selectedSensors []api.Sensor, data api.Data, apiConfig config.Config) Export {
	name := getFileName(apiConfig)

	return Export{
		energyTypes:     apiConfig.EnergyTypes,
		samples:         samples,
		data:            data,
		selectedSensors: selectedSensors,
		fileName:        name,
		fileType:		 apiConfig.Format,
	}
}

func (e *Export) Write() {
	if e.fileType == "json" {
		e.JSON()
	}

	if e.fileType == "csv" {
		e.CSV()
	}
}

func (e *Export) JSON() {
	file, err := os.Create(e.fileName)
	if err != nil {
		log.Fatalln("error creating file", err)
	}

	defer file.Close()

	samplesJson, _ := json.Marshal(e.samples)
	file.Write(samplesJson)
	log.Printf("Created file: %s\n", e.fileName)
}

func (e *Export) CSV() {
	file, err := os.Create(e.fileName)
	if err != nil {
		log.Fatalln("error creating file", err)
	}

	defer file.Close()
	w := csv.NewWriter(file)

	// timestamp, sensorIdA power, sensorIdA energy, sensorIdB power, sensorIdB energy
	csvHeader := []string{"timestamp"}
	for _, sensor := range e.selectedSensors {
		for _, energyType := range e.energyTypes {
			csvHeader = append(csvHeader, fmt.Sprintf("%s %s", sensor.Id, energyType))
		}
	}

	if err := w.Write(csvHeader); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	for _, column := range e.data {
		var csvRow []string
		csvRow = append(csvRow, column.DateTime.UTC().Format("2006-01-02 15:04:05"))
		for _, sensor := range e.selectedSensors {
			for _, energyType := range e.energyTypes {
				v := column.Readings[sensor.Id][energyType]
				if v == nil {
					csvRow = append(csvRow, "")
				} else {
					csvRow = append(csvRow, strconv.FormatFloat(*v, 'f', 8, 64))
				}
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
}

func getFileName(apiConfig config.Config) string {
	energyTypes := strings.Join(apiConfig.EnergyTypes, "-")
	return fmt.Sprintf("%d_%d_%s_%s_%s.%s", apiConfig.TimeFrom.Unix(), apiConfig.TimeTo.Unix(), apiConfig.DataLogger, energyTypes, apiConfig.Aggregation.Level, apiConfig.Format)
}
