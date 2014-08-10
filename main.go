package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/template"
	"time"
)

type Temp float64

func (t Temp) ToC() float64 {
	return float64(t) - 273.15
}

func (t Temp) ToF() float64 {
	return t.ToC()*1.8 + 32
}

func FmtTemp(f float64) string {
	return strconv.FormatFloat(f, 'f', 1, 64)
}

type CurrentWeather struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Sys struct {
		Message float64 `json:"message"`
		Country string  `json:"country"`
		Sunrise int64   `json:"sunrise"`
		Sunset  int64   `json:"sunset"`
	} `json:"sys"`
	Weather []struct {
		Id          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		// Unit is K
		Temp Temp `json:"temp"`
		// Unit is K
		TempMin Temp `json:"temp_min"`
		// Unit is K
		TempMax Temp `json:"temp_max"`
		// Unit is hpa
		Pressure    float64 `json:"pressure"`
		SeaLevel    float64 `json:"sea_level"`
		GroundLevel float64 `json:"grnd_level"`
		Humidity    int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Rain struct {
		// mm/3 hour of rain
		ThreeHours float64 `json:"3h"`
		OneHour    float64 `json:"1h"`
	} `json:"rain"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt   int64  `json:"dt"`
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Cod  int    `json:"cod"`
}

const API_URL = "http://api.openweathermap.org/data/2.5/weather"

func FetchWeather(cityId string) (*CurrentWeather, error) {
	u, err := url.Parse(API_URL)
	if err != nil {
		return nil, err
	}
	query := url.Values{"id": {cityId}}
	u.RawQuery = query.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to retrieve weather (HTTP %d)", resp.StatusCode)
	}

	var cw CurrentWeather
	if err := json.NewDecoder(resp.Body).Decode(&cw); err != nil {
		return nil, err
	}
	return &cw, err
}

func HandleWeather(w io.Writer, cityId string, tpl *template.Template, updateCh <-chan bool, updateCounterCh chan<- bool) {
	for _ = range updateCh {
		currentWeather, err := FetchWeather(cityId)
		if err != nil {
			log.Println(err)
			return
		}
		if err := tpl.Execute(w, currentWeather); err != nil {
			log.Println(err)
			return
		}
		updateCounterCh <- true
	}
}

func main() {
	var fetchDelay time.Duration
	flag.DurationVar(&fetchDelay, "fetch-delay", time.Minute*30, "How much time between each fetch.")
	var runOnce bool
	flag.BoolVar(&runOnce, "once", false, "Run once and exit")

	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		log.Fatal("Missing City ID")
	}

	// Get city ID
	cityId := flag.Arg(0)

	// Parse template
	tplBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	tplFuncMap := template.FuncMap{
		"temp": FmtTemp,
	}
	weatherTemplate, err := template.New("weather").Funcs(tplFuncMap).Parse(string(tplBytes))
	if err != nil {
		log.Fatal(err)
	}

	updateCh := make(chan bool)
	updateCounterCh := make(chan bool)
	go HandleWeather(os.Stdout, cityId, weatherTemplate, updateCh, updateCounterCh)

	go func() {
		count := 0
		for _ = range updateCounterCh {
			count++
			if count == 1 && runOnce {
				os.Exit(0)
			}
		}
	}()
	updateCh <- true

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	// Tick, fetch and format weather
	ticks := time.Tick(fetchDelay)
	// Update now
	for {
		select {
		case <-ticks:
			updateCh <- true
		case <-sigs:
			updateCh <- true
		}
	}
}
