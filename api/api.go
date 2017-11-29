// Package api provides the api communication
package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/wiesson/eb-export/config"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const baseUrl = "https://api.internetofefficiency.com"

type api struct {
	config config.Config
	Data   Data
}

// New returns an instance of api
func New(config config.Config) api {
	return api{config: config}
}

type response struct {
	Data  []responseData `json:"data"`
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

type Sample struct {
	Timestamp int64
	DateTime  time.Time
	Readings  map[string]*float64
}

type Data []Sample

func (d *Data) addReading(value responseData, energyType string) {
	DateTime := time.Unix(value.Attributes.Timestamp, 0)

	row := &Sample{
		Timestamp: value.Attributes.Timestamp,
		DateTime:  DateTime,
		Readings:  make(map[string]*float64),
	}

	switch energyType {
	case "energy":
		for _, sample := range value.Attributes.EnergyResponseSamples {
			row.Readings[sample.SensorID] = &sample.Value
		}
	default:
		for _, sample := range value.Attributes.PowerResponseSamples {
			row.Readings[sample.SensorID] = &sample.Value
		}
	}

	*d = append(*d, *row)
}

func (a *api) Fetch() []Sample {
	nextUrl := a.getRequestPath("")
	hasNext := true

	for hasNext {
		res, err := a.get(nextUrl)
		if err != nil {
			log.Fatal(err)
		}

		logMessage, err := url.ParseQuery(nextUrl)
		offset := logMessage.Get("page[offset]")

		if offset != "" {
			log.Printf("Fetching from %s\n", logMessage.Get("page[offset]"))
		}

		for _, value := range res.Data {
			a.Data.addReading(value, a.config.EnergyType)
		}

		nextUrl = res.Links.NextURL
		if nextUrl == "" {
			hasNext = false
			break
		}
	}

	return a.Data
}

func (a *api) get(url string) (response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseUrl+url, nil)
	if err != nil {
		return response{}, fmt.Errorf("unable to create new request: %v", err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.config.AccessToken))

	res, err := client.Do(req)
	if err != nil {
		return response{}, fmt.Errorf("unable to do API request: %v", err)
	}
	defer res.Body.Close()

	s := &response{}

	body, err := gzip.NewReader(res.Body)
	if err != nil {
		return response{}, fmt.Errorf("unable to decode gzipped resonse: %v", err)
	}

	err = json.NewDecoder(body).Decode(s)
	if err != nil {
		return response{}, fmt.Errorf("unable to parse JSON: %v", err)
	}

	return *s, nil
}

func (a *api) getRequestPath(path string) string {
	if path != "" {
		return path
	}

	payload := url.Values{}
	payload.Set("aggregation_level", a.config.AggregationLevel)
	payload.Add("filter[samples]", fmt.Sprintf("timestamp,%s", a.config.EnergyType))
	payload.Add("filter[from]", strconv.FormatInt(a.config.TimeFrom, 10))
	payload.Add("filter[to]", strconv.FormatInt(a.config.TimeTo, 10))
	payload.Add("filter[data_logger]", a.config.DataLogger)
	payload.Add("filter[sensor]", strings.Join(a.config.Sensors, ","))
	return "/v2/samples/?" + payload.Encode()
}
