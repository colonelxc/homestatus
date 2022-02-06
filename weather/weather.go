// This package gets forcast data from api.weather.gov
// weather.gov breaks everything into 'grid points'.
// You need to use api.weather.gov/points/<lat>,<long> to figure out the correct grid points
// this package does not understand json-ld generically, just this particular API endpoint.
package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

// Forecast periods are returned in order.
func GetForecast(gridId string, gridX, gridY int) ([]*ForecastPeriod, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.weather.gov/gridpoints/%s/%d,%d/forecast", gridId, gridX, gridY), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "HomeStatus/1.0 (http://github.com/colonelxc/homestatus)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

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
	return periods, nil
}
