// Package api provides the api communication
package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/wiesson/eb-export/config"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const baseUrl = "https://api.internetofefficiency.com"

type api struct {
	config  config.Config
	Sensors []Sensor
	Logger
	Samples
	Data
}

// New returns an instance of api
func New(config config.Config) api {
	return api{config: config}
}

type Logger struct {
	Id              string
	Description     string
	Building        string
	MacAddress      string
	SampleFrequency int64
	NumPhases       int64
	Mdp             bool
	CreatedAt       int64
}

type loggerResponse struct {
	Data responseLoggerData `json:"data"`
}

type responseLoggerData struct {
	Type       string `json:"type"`
	Id         string `json:"id"`
	Attributes struct {
		Description     string   `json:"description"`
		Building        string   `json:"building"`
		MacAddress      string   `json:"mac_address"`
		SampleFrequency int64    `json:"sample_frequency"`
		NumPhases       int64    `json:"num_phases"`
		Mdp             bool     `json:"mdp"`
		CreatedAt       int64    `json:"created_at"`
		Sensors         []Sensor `json:"sensors"`
	}
}

type Sensor struct {
	Id             string `json:"sensor_id"`
	Type           string `json:"type"`
	Phase          int64  `json:"phase"`
	Description    string `json:"description"`
	BuildingFloor  string `json:"building_floor"`
	FunctionalArea string `json:"functional_area"`
	Room           string `json:"room"`
	EquipmentGroup string `json:"equipment_group"`
	EquipmentType  string `json:"equipment_type"`
}

type samplesResponse struct {
	Data  []responseSampleData `json:"data"`
	Links struct {
		NextURL string `json:"next"`
	} `json:"links"`
}

type responseSampleData struct {
	Type       string `json:"type"`
	Id         string `json:"id"`
	Attributes struct {
		Timestamp              int64            `json:"timestamp"`
		PowerResponseSamples   []responseSample `json:"power"`
		EnergyResponseSamples  []responseSample `json:"energy"`
		CurrentResponseSamples []responseSample `json:"current"`
	} `json:"attributes"`
}

type responseSample struct {
	Id    string  `json:"sensor_id"`
	Value float64 `json:"value"`
}

type Sample struct {
	Timestamp int64
	DateTime  time.Time
	Readings  map[string]map[string]*float64
}

type Samples []responseSampleData

func (s *Samples) Add(data responseSampleData) {
	*s = append(*s, data)
}

type Data []Sample

func (d *Data) addReading(value responseSampleData, sensors []Sensor, energyTypes []string) {
	DateTime := time.Unix(value.Attributes.Timestamp, 0)

	var readings = map[string]map[string]*float64{}
	for _, sensor := range sensors {
		readings[sensor.Id] = make(map[string]*float64)
	}

	row := &Sample{
		Timestamp: value.Attributes.Timestamp,
		DateTime:  DateTime,
		Readings:  readings,
	}

	for _, energyType := range energyTypes {
		switch energyType {
		case "power":
			for _, sample := range value.Attributes.PowerResponseSamples {
				var v float64
				v = sample.Value
				row.Readings[sample.Id][energyType] = &v
			}

		case "energy":
			for _, sample := range value.Attributes.EnergyResponseSamples {
				var v float64
				v = sample.Value
				row.Readings[sample.Id][energyType] = &v
			}

		case "current":
			for _, sample := range value.Attributes.CurrentResponseSamples {
				var v float64
				v = sample.Value
				row.Readings[sample.Id][energyType] = &v
			}
		}
	}

	*d = append(*d, *row)
}

func (a *api) FetchLogger() []Sensor {
	loggerResponse, err := a.getLogger()

	if err != nil {
		log.Fatal(err.Error())
	}

	a.Logger.Id = loggerResponse.Data.Id
	a.Logger.Description = loggerResponse.Data.Attributes.Description
	a.Logger.Building = loggerResponse.Data.Attributes.Building
	a.Logger.MacAddress = loggerResponse.Data.Attributes.MacAddress
	a.Logger.SampleFrequency = loggerResponse.Data.Attributes.SampleFrequency
	a.Logger.NumPhases = loggerResponse.Data.Attributes.NumPhases
	a.Logger.Mdp = loggerResponse.Data.Attributes.Mdp
	a.Logger.CreatedAt = loggerResponse.Data.Attributes.CreatedAt
	a.Sensors = loggerResponse.Data.Attributes.Sensors
	return a.Sensors
}

func (a *api) FetchSamples() ([]Sensor, Samples, Data) {
	var selectedSensors []Sensor
	if len(a.config.InputSensors) > 0 {
		for _, sensor := range a.Sensors {
			for _, selectedSensor := range a.config.InputSensors {
				if sensor.Id == selectedSensor {
					selectedSensors = append(selectedSensors, sensor)
				}
			}
		}
	} else {
		selectedSensors = a.Sensors
	}

	// split range into single days
	for d := a.config.TimeFrom; d.Before(a.config.TimeTo); d = d.AddDate(0, 0, 1) {
		start := d
		end := start.AddDate(0, 0, 1)

		log.Printf("Fetching from %s to %s\n", start, end)

		nextUrl := a.getSamplesParameters("", start, end)
		hasNext := true

		for hasNext {
			res, err := a.getSamples(nextUrl)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Received %d", len(res.Data))

			logMessage, err := url.ParseQuery(nextUrl)
			offset := logMessage.Get("page[offset]")

			if offset != "" {
				log.Printf("Fetching from offset %s\n", logMessage.Get("page[offset]"))
			}

			for _, value := range res.Data {
				if a.config.Format == "json" {
					a.Samples.Add(value)
				} else {
					a.Data.addReading(value, selectedSensors, a.config.EnergyTypes)
				}
			}

			nextUrl = res.Links.NextURL
			if nextUrl == "" {
				hasNext = false
				break
			}
		}
	}

	return selectedSensors, a.Samples, a.Data
}

func (a *api) getLogger() (loggerResponse, error) {
	var client http.Client
	req, _ := a.NewGetRequest("/v2/data_loggers/" + a.config.DataLogger)
	res, err := client.Do(req)
	if err != nil {
		return loggerResponse{}, err
	}
	defer res.Body.Close()

	var body io.ReadCloser

	if res.StatusCode == 401 {
		return loggerResponse{}, fmt.Errorf("token expired or not authorized")
	}

	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		body, err = gzip.NewReader(res.Body)
		if err != nil {
			return loggerResponse{}, fmt.Errorf("unable to decode gzipped response: %v", err)
		}
		defer body.Close()
	default:
		body = res.Body
	}

	payload := &loggerResponse{}
	err = json.NewDecoder(body).Decode(payload)
	if err != nil {
		return loggerResponse{}, fmt.Errorf("unable to parse JSON: %v", err)
	}

	body = nil

	return *payload, nil
}

func (a *api) getSamples(requestUrl string) (samplesResponse, error) {
	var client http.Client
	req, _ := a.NewGetRequest(requestUrl)
	res, err := client.Do(req)
	if err != nil {
		return samplesResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		return samplesResponse{}, fmt.Errorf("token expired or not authorized")
	}

	var body io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		body, err = gzip.NewReader(res.Body)
		if err != nil {
			return samplesResponse{}, fmt.Errorf("unable to decode gzipped response: %v", err)
		}
		defer body.Close()
	default:
		body = res.Body
	}

	payload := &samplesResponse{}
	err = json.NewDecoder(body).Decode(payload)
	if err != nil {
		return samplesResponse{}, fmt.Errorf("unable to parse http response: %v", err)
	}

	return *payload, nil
}

func (a *api) getSamplesParameters(path string, start, end time.Time) string {
	if path != "" {
		return path
	}

	fields := []string{"timestamp", strings.Join(a.config.EnergyTypes, ",")}

	payload := url.Values{}
	payload.Set("aggregation_level", a.config.Aggregation.Level)
	payload.Add("filter[from]", strconv.FormatInt(start.Unix(), 10))
	payload.Add("filter[to]", strconv.FormatInt(end.Unix(), 10))
	payload.Add("filter[data_logger]", a.config.DataLogger)
	payload.Add("fields[samples]", strings.Join(fields, ","))

	if len(a.config.InputSensors) > 0 {
		payload.Add("filter[sensor]", strings.Join(a.config.InputSensors, ","))
	}

	return "/v2/samples?" + payload.Encode()
}

func (a *api) NewGetRequest(requestUrl string) (*http.Request, error) {
	req, err := http.NewRequest("GET", baseUrl+requestUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %v", err)
	}

	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.config.AccessToken))

	return req, nil
}
