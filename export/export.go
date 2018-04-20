package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/wiesson/eb-export/api"
	"github.com/wiesson/eb-export/config"
	"log"
	"os"
	"strconv"
	"strings"
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
		fileType:        apiConfig.Format,
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
	csvHeaderEnergyType := []string{"timestamp"}
	csvHeaderSensorId := []string{"id"}
	csvHeaderDescription := []string{"description"}
	csvHeaderFunctionalArea := []string{"functional_area"}
	csvHeaderRoom := []string{"room"}

	for _, sensor := range e.selectedSensors {
		for _, energyType := range e.energyTypes {
			csvHeaderEnergyType = append(csvHeaderEnergyType, fmt.Sprintf("%s", energyType))
		}

		csvHeaderSensorId = append(csvHeaderSensorId, fmt.Sprintf("%s", sensor.Id))
		csvHeaderDescription = append(csvHeaderDescription, fmt.Sprintf("%s", sensor.Description))
		csvHeaderFunctionalArea = append(csvHeaderFunctionalArea, fmt.Sprintf("%s", sensor.FunctionalArea))
		csvHeaderRoom = append(csvHeaderRoom, fmt.Sprintf("%s", sensor.Room))

		for i := 0; i < len(e.energyTypes)-1; i++ {
			csvHeaderSensorId = append(csvHeaderSensorId, fmt.Sprint(""))
			csvHeaderDescription = append(csvHeaderDescription, fmt.Sprint(""))
			csvHeaderFunctionalArea = append(csvHeaderFunctionalArea, fmt.Sprint(""))
			csvHeaderRoom = append(csvHeaderRoom, fmt.Sprint(""))
		}
	}

	if err := w.Write(csvHeaderSensorId); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	if err := w.Write(csvHeaderDescription); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	if err := w.Write(csvHeaderFunctionalArea); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	if err := w.Write(csvHeaderRoom); err != nil {
		log.Fatalln("error writing header to csv:", err)
	}

	if err := w.Write(csvHeaderEnergyType); err != nil {
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
