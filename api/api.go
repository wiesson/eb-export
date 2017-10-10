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

type Response struct {
	Data []ResponseData `json:"data"`
	Links struct {
		NextURL string `json:"next"`
	} `json:"links"`
}

type ResponseData struct {
	Type       string `json:"type"`
	Id         string `json:"id"`
	Attributes struct {
		Timestamp             int64            `json:"timestamp"`
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

type Sample struct {
	Timestamp int64
	DateTime  time.Time
	Readings  map[string]Reading
}

type Data []Sample

func (d *Data) AddItem(value ResponseData, energyType string) {
	DateTime := time.Unix(value.Attributes.Timestamp, 0)

	row := &Sample{
		Timestamp: value.Attributes.Timestamp,
		DateTime:  DateTime,
		Readings:  make(map[string]Reading),
	}

	if energyType == "power" {
		for _, sample := range value.Attributes.PowerResponseSamples {
			row.Readings[sample.SensorID] = Reading(sample.Value)
		}
	}

	if energyType == "energy" {
		for _, sample := range value.Attributes.EnergyResponseSamples {
			row.Readings[sample.SensorID] = Reading(sample.Value)
		}
	}

	*d = append(*d, *row)
}


func (a *api) Get(url string) (Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseUrl+url, nil)
	if err != nil {
		return Response{}, fmt.Errorf("unable to create new request: %v", err)
	}
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("unable to do API request: %v", err)
	}
	defer res.Body.Close()

	s := &Response{}

	body, err :=  gzip.NewReader(res.Body)
	if err != nil {
		return Response{}, fmt.Errorf("unable to decode gzipped resonse: %v", err)
	}

	err = json.NewDecoder(body).Decode(s)
	if err != nil {
		return Response{}, fmt.Errorf("unable to parse JSON: %v", err)
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
