package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/colonelxc/homestatus/serialize"
	"github.com/colonelxc/homestatus/weather"
)

var (
	lat  = flag.String("latitude", "", "Latitude for weather data")
	long = flag.String("longitude", "", "Longitude for weather data")
	port = flag.Int("port", 8080, "port to run on (listens on localhost)")
)

type data struct {
	updateTime time.Time
	periods    []*weather.ForecastPeriod
}

var (
	mostRecentData atomic.Value
)

func main() {
	flag.Parse()
	if *lat == "" || *long == "" {
		fmt.Println("latitude and longitude required")
		return
	}

	forecastUrl, err := weather.GetForecastUrl(*lat, *long)
	if err != nil {
		log.Fatalf("Could not get the forecast url: %w", err)
	}

	doUpdate(time.Now(), forecastUrl)
	go backgroundUpdater(time.Hour, forecastUrl)
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/data", handleData)
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", *port), nil)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q. This service is for a dashboard, not you!", html.EscapeString(r.URL.Path))
}

func handleData(w http.ResponseWriter, r *http.Request) {
	origin := r.RemoteAddr
	if f := r.Header.Get("X-Forwarded-For"); f != "" {
		origin = origin + ", X-Forwarded-For=" + f
	}
	log.Printf("Handling data request from: %s", origin)
	recentData, ok := mostRecentData.Load().(data)
	if !ok {
		errorResponse(w, 500, "Error getting the forecast")
		return
	}

	b := new(bytes.Buffer)

	s := serialize.NewWriter(b)
	writeUpdateTime(s, recentData.updateTime)
	writeForecast(s, recentData.periods)
	s.Finish()
	if s.Err() != nil {
		log.Printf("Error writing out the forecast: %s", s.Err())
		errorResponse(w, 500, "Error writing the forecast")
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.Write(b.Bytes())
}

func writeUpdateTime(s *serialize.Serializer, lastUpdated time.Time) {
	s.NextDataType("UpdateTime")
	s.WriteColumnNames([]string{"lastUpdated", "currentTime", "secondsToNextUpdate"})
	last := lastUpdated.Local().Format(time.RFC1123)
	now := time.Now().Local().Format(time.RFC1123)
	s.AddRow().WriteStringValue(last).WriteStringValue(now).WriteIntValue(60 * 60 /* 1 hour */).Done()
}

func writeForecast(s *serialize.Serializer, periods []*weather.ForecastPeriod) {
	s.NextDataType("WeatherForecast")
	s.WriteColumnNames([]string{"Name", "IsDayTime", "Temperature", "WindSpeed", "WindDirection", "ShortForecast"})
	for _, p := range periods {
		s.AddRow().WriteStringValue(p.Name).WriteBoolValue(p.IsDayTime).WriteStringValue(p.Temperature).WriteStringValue(p.WindSpeed).WriteStringValue(p.WindDirection).WriteStringValue(p.ShortForecast).Done()
	}
}

func errorResponse(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	fmt.Fprintln(w, msg)
}

func backgroundUpdater(d time.Duration, forecastUrl string) {
	ticker := time.NewTicker(d)
	for t := range ticker.C {
		doUpdate(t, forecastUrl)
	}
}

func doUpdate(t time.Time, forecastUrl string) {
	log.Println("Performing data update")
	periods, err := weather.GetForecast(forecastUrl)
	if err != nil {
		log.Printf("Error getting the forecast %s", err)
		return
	}
	mostRecentData.Store(data{
		updateTime: t,
		periods:    periods,
	})
}
