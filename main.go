package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

type SamplesResponse struct {
	Sample []SamplesResponseData `json:"data"`
	Meta struct {
		SampleInterval uint `json:"sample_interval"`
	} `json:"meta"`
	Links struct {
		NextURL string `json:"next"`
	} `json:"links"`
}

type SamplesResponseData struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Attributes struct {
		Timestamp             int64            `json:"timestamp"`
		SystemTemperature     float32          `json:"system_temperature"`
		PowerResponseSamples  []ResponseSample `json:"power"`
		EnergyResponseSamples []ResponseSample `json:"energy"`
	} `json:"attributes"`
}

type ResponseSample struct {
	SensorID string  `json:"sensor_id"`
	Value    float64 `json:"value"`
}

type Reading float64

func (r Reading) String() string {
	return strconv.FormatFloat(float64(r), 'f', 8, 64)
}

type Data []Sample

func (d *Data) AddItem(value SamplesResponseData, energyType string) {
	DateTime := time.Unix(value.Attributes.Timestamp, 0)

	row := &Sample{
		Timestamp: value.Attributes.Timestamp,
		DateTime:  DateTime,
		Samples: make(map[string]Reading),
	}

	if energyType == "power" {
		for _, sample := range value.Attributes.PowerResponseSamples {
			row.Samples[sample.SensorID] = Reading(sample.Value)
		}
	}

	if energyType == "energy" {
		for _, sample := range value.Attributes.EnergyResponseSamples {
			row.Samples[sample.SensorID] = Reading(sample.Value)
		}
	}

	*d = append(*d, *row)
}

type Sample struct {
	Timestamp int64
	DateTime  time.Time
	Samples map[string]Reading
}

type API struct {
	baseUrl          string
	dataLogger       string
	sensors          argsWithSameKey
	timeFrom         int64
	timeTo           int64
	aggregationLevel string
	tz               time.Location
	energyType       string
}

func (a *API) Get(url string) (SamplesResponse, error) {
	res, err := http.Get(a.baseUrl + url)
	defer res.Body.Close()
	if err != nil {
		return SamplesResponse{}, err
	}

	s := &SamplesResponse{}
	err = json.NewDecoder(res.Body).Decode(s)
	if err != nil {
		return SamplesResponse{}, err
	}

	return *s, nil
}

func (a *API) GetRequestPath (path string) string {
	if path != "" {
		return path
	}

	payload := url.Values{}
	payload.Set("aggregation_level", a.aggregationLevel)
	payload.Add("filter[samples]", fmt.Sprintf("timestamp,%s", a.energyType))
	payload.Add("filter[from]", strconv.FormatInt(a.timeFrom, 10))
	payload.Add("filter[to]", strconv.FormatInt(a.timeTo, 10))
	payload.Add("filter[data_logger]", a.dataLogger)
	payload.Add("filter[sensor]", a.sensors.String())
	return "/v2/samples/?" + payload.Encode()
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

	api := API{
		baseUrl:          "https://api.internetofefficiency.com",
		dataLogger:       *logger,
		sensors:          sensors,
		timeFrom:         lower.Unix(),
		timeTo:           upper.Unix(),
		aggregationLevel: *aggregationLevel,
		tz:               *loc,
		energyType:       * energyType,
	}

	d := &Data{}

	fmt.Print("Fetching data")

	nextUrl := api.GetRequestPath("")
	hasNext := true

	for hasNext {
		s, err := api.Get(nextUrl)
		if err != nil {
			panic(err)
		}
		for _, value := range s.Sample {
			d.AddItem(value, *energyType)
		}
		NewNextUrl := s.Links.NextURL
		if NewNextUrl == "" {
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

	for _, column := range *d {
		var csvRow []string
		csvRow = append(csvRow, column.DateTime.UTC().Format("2006-01-02 15:04:05")) // time.RFC3339 // 2006-01-02T15:04:05.999
		for _, sensorID := range sensors {
			v := column.Samples[sensorID]
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
