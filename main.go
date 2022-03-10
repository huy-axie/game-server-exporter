package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Data struct {
	Total int `json:"total"`
}

var battles int

var (
	battlesNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_battle_total",
		Help:        "The total number of match happening",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})
)

// set metrics
func recordMetrics() {
	go func() {
		for {
			// get battles
			getBattles()
			// set metrics
			battlesNumber.Set(float64(battles))
			time.Sleep(15 * time.Second)
		}
	}()
}

func getBattles() {
	httpUrl := "http://" + os.Getenv("GAME_SERVER_IP") + ":" + os.Getenv("GAME_SERVER_PORT") + "/" + os.Getenv("BATTLE_PATH")
	fmt.Println("Open connection to " + httpUrl)

	req, err := http.NewRequest("GET", httpUrl, nil)
	if err != nil {
		panic(err)
	}

	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{os.Getenv("GAME_SERVER_TOKEN")},
	}
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	dat := Data{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}

	battles = dat.Total
}

// Get node name
func getHostName() string {
	name, err := os.Hostname()
	if err != nil {
		panic("Can't get hostname")
	}
	return name
}

func main() {
	// register metrics
	prometheus.MustRegister(battlesNumber)

	// record metrics
	recordMetrics()

	// start serving http metrics
	fmt.Println("Start metrics at :9101/metrics")
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9101", nil))
}
