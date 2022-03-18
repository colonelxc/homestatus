package main

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/colonelxc/homestatus/serialize"
	"github.com/colonelxc/homestatus/weather"
)

type data struct {
	updateTime time.Time
	periods    []*weather.ForecastPeriod
}

var mostRecentData atomic.Value

func main() {
	doUpdate(time.Now())
	go backgroundUpdater(time.Hour)
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/data", handleData)
	http.ListenAndServe("127.0.0.1:8080", nil)
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

func backgroundUpdater(d time.Duration) {
	ticker := time.NewTicker(d)
	for t := range ticker.C {
		doUpdate(t)
	}
}

func doUpdate(t time.Time) {
	log.Println("Performing data update")
	periods, err := weather.GetForecast("SEW", 128, 69)
	if err != nil {
		log.Printf("Error getting the forecast %s", err)
		return
	}
	mostRecentData.Store(data{
		updateTime: t,
		periods:    periods,
	})
}
