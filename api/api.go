package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"compress/gzip"
	"log"
)

const baseUrl = "https://api.internetofefficiency.com"


type api struct {
	DataLogger       string
	Sensors          []string
	TimeFrom         int64
	TimeTo           int64
	AggregationLevel string
	Tz               time.Location
	EnergyType       string
	Data 			 data
}

func Config(DataLogger string, sensors []string, timeFrom int64, timeTo int64, aggregationLevel string, timezone time.Location, energyType string) api {
	return api{
		DataLogger:       DataLogger,
		Sensors:          sensors,
		TimeFrom:         timeFrom,
		TimeTo:           timeTo,
		AggregationLevel: aggregationLevel,
		Tz:               timezone,
		EnergyType:       energyType,
	}
}

type response struct {
	Data []responseData `json:"data"`
	Links struct {
		NextURL string `json:"next"`
	} `json:"links"`
}

type responseData struct {
	Type       string `json:"type"`
	Id         string `json:"id"`
	Attributes struct {
		Timestamp             int64            `json:"timestamp"`
		PowerResponseSamples  []responseSample `json:"power"`
		EnergyResponseSamples []responseSample `json:"energy"`
	} `json:"attributes"`
}

type responseSample struct {
	SensorID string  `json:"sensor_id"`
	Value    float64 `json:"value"`
}

type Reading float64

func (r Reading) String() string {
	return strconv.FormatFloat(float64(r), 'f', 8, 64)
}

type Sample struct {
	Timestamp int64
	DateTime  time.Time
	Readings  map[string]Reading
}

type data []Sample

func (d *data) AddItem(value responseData, energyType string) {
	DateTime := time.Unix(value.Attributes.Timestamp, 0)

	row := &Sample{
		Timestamp: value.Attributes.Timestamp,
		DateTime:  DateTime,
		Readings:  make(map[string]Reading),
	}

	switch energyType {
	case "energy":
		for _, sample := range value.Attributes.EnergyResponseSamples {
			row.Readings[sample.SensorID] = Reading(sample.Value)
		}
	default:
		for _, sample := range value.Attributes.PowerResponseSamples {
			row.Readings[sample.SensorID] = Reading(sample.Value)
		}
	}

	*d = append(*d, *row)
}

func (a *api) Fetch() []Sample {
	nextUrl := a.GetRequestPath("")
	hasNext := true

	for hasNext {
		res, err := a.Get(nextUrl)
		if err != nil {
			log.Fatal(err)
		}

		logMessage, err := url.ParseQuery(nextUrl)
		log.Printf("Fetching from %s\n", logMessage.Get("page[offset]"))

		for _, value := range res.Data {
			a.Data.AddItem(value, a.EnergyType)
		}

		nextUrl = res.Links.NextURL
		if nextUrl == "" {
			hasNext = false
			break
		}
	}

	return a.Data
}

func (a *api) Get(url string) (response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseUrl+url, nil)
	if err != nil {
		return response{}, fmt.Errorf("unable to create new request: %v", err)
	}
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := client.Do(req)
	if err != nil {
		return response{}, fmt.Errorf("unable to do API request: %v", err)
	}
	defer res.Body.Close()

	s := &response{}

	body, err :=  gzip.NewReader(res.Body)
	if err != nil {
		return response{}, fmt.Errorf("unable to decode gzipped resonse: %v", err)
	}

	err = json.NewDecoder(body).Decode(s)
	if err != nil {
		return response{}, fmt.Errorf("unable to parse JSON: %v", err)
	}

	return *s, nil
}

func (a *api) GetRequestPath(path string) string {
	if path != "" {
		return path
	}

	payload := url.Values{}
	payload.Set("aggregation_level", a.AggregationLevel)
	payload.Add("filter[samples]", fmt.Sprintf("timestamp,%s", a.EnergyType))
	payload.Add("filter[from]", strconv.FormatInt(a.TimeFrom, 10))
	payload.Add("filter[to]", strconv.FormatInt(a.TimeTo, 10))
	payload.Add("filter[data_logger]", a.DataLogger)
	payload.Add("filter[sensor]", strings.Join(a.Sensors, ","))
	return "/v2/samples/?" + payload.Encode()
}
