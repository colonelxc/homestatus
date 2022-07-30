// This package gets forcast data from api.weather.gov
// weather.gov breaks everything into 'grid points'.
// You need to use api.weather.gov/points/<lat>,<long> to figure out the correct grid points
// this package does not understand json-ld generically, just this particular API endpoint.
package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// minimal type for extracting forecast url from lat/long metadata query
type point struct {
	Properties struct {
		Forecast string `json:"forecast"`
	} `json:"properties"`
}

// internal forcast type, for parsing from json
type forecast struct {
	Properties properties `json:"properties"`
}

type properties struct {
	Periods []period `json:"Periods"`
}

type period struct {
	Name string `json:"name"` // ex: "This Afternoon"
	//	startTime        string // ex: "2022-02-04T13:00:00-08:00"
	//  endTime          string // ex: "2022-02-04T18:00:00-08:00"
	IsDayTime       bool   `json:"isDayTime"`       // ex: true
	Temperature     int    `json:"temperature"`     // ex: 48
	TemperatureUnit string `json:"temperatureUnit"` // ex: "F"
	WindSpeed       string `json:"windSpeed"`       // ex: "12 mph" or "1 to 6 mph"
	WindDirection   string `json:"windDirection"`   // ex: "SSW"
	ShortForecast   string `json:"shortForecast"`   // ex: "Light Rain Likely"
	// detailedForecast string // ex: "Rain likely. Mostly cloudy, with a high near 48. South southwest wind around 12 mph. Chance of precipitation is 70%. New rainfall amounts less than a tenth of an inch possible."
}

type ForecastPeriod struct {
	Name          string
	IsDayTime     bool
	Temperature   string // ex: 48F
	WindSpeed     string // ex: "12 mph"
	WindDirection string // ex: "SSW"
	ShortForecast string // ex: "Light Rain Likely"
}

func GetForecastUrl(lat, long string) (string, error) {
	decoder, closer, err := doRequest(fmt.Sprintf("https://api.weather.gov/points/%s,%s", lat, long))
	if err != nil {
		return "", err
	}
	defer closer()

	p := point{}
	err = decoder.Decode(&p)
	if err != nil {
		return "", fmt.Errorf("error reading forecast points json: %w", err)
	}
	return p.Properties.Forecast, nil
}

// Forecast periods are returned in order.
func GetForecast(forecastUrl string) ([]*ForecastPeriod, error) {
	decoder, closer, err := doRequest(forecastUrl)
	if err != nil {
		return nil, err
	}
	defer closer()

	f := forecast{}
	err = decoder.Decode(&f)
	if err != nil {
		return nil, err
	}

	var periods []*ForecastPeriod
	for _, v := range f.Properties.Periods {
		periods = append(periods, &ForecastPeriod{
			Name:          v.Name,
			IsDayTime:     v.IsDayTime,
			Temperature:   fmt.Sprintf("%d%s", v.Temperature, v.TemperatureUnit),
			WindSpeed:     v.WindSpeed,
			WindDirection: v.WindDirection,
			ShortForecast: v.ShortForecast,
		})
	}
	if len(periods) == 0 {
		return nil, errors.New("received no forecast periods from the weather API")
	}
	return periods, nil
}

func doRequest(url string) (*json.Decoder, func() error, error) {
	log.Printf("Requesting url: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "HomeStatus/1.0 (http://github.com/colonelxc/homestatus)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("weather api error: %s", resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	return decoder, resp.Body.Close, nil
}
