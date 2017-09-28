package samples

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

func New() {}

type API struct {
	BaseUrl          string
	DataLogger       string
	Sensors          []string
	TimeFrom         int64
	TimeTo           int64
	AggregationLevel string
	Tz               time.Location
	EnergyType       string
}

type SamplesResponse struct {
	Sample []SamplesResponseData `json:"data"`
	Meta   struct {
		SampleInterval uint `json:"sample_interval"`
	} `json:"meta"`
	Links struct {
		NextURL string `json:"next"`
	} `json:"links"`
}

type SamplesResponseData struct {
	Type       string `json:"type"`
	Id         string `json:"id"`
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
		Samples:   make(map[string]Reading),
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
	Samples   map[string]Reading
}

func (a *API) Get(url string) (SamplesResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", a.BaseUrl+url, nil)
	if err != nil {
		return SamplesResponse{}, err
	}
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := client.Do(req)
	if err != nil {
		return SamplesResponse{}, err
	}
	defer res.Body.Close()

	s := &SamplesResponse{}

	body, err :=  gzip.NewReader(res.Body)
	if err != nil {
		return SamplesResponse{}, err
	}

	err = json.NewDecoder(body).Decode(s)
	if err != nil {
		return SamplesResponse{}, err
	}

	return *s, nil
}

func (a *API) GetRequestPath(path string) string {
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
