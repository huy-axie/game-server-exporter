package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const bitSize = 64

var battles float64
var done chan interface{}
var interrupt chan os.Signal
var (
	battlesNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "game_server_battle_total",
		Help: "The total number of match happening",
	})
)

var (
	hostName = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "game_server_hostname",
			Help:        "Game server hostname",
			ConstLabels: map[string]string{"nodename": getHostName()},
		})
)

func recordMetrics() {
	go func() {
		for {
			battlesNumber.Set(battles)
			time.Sleep(30 * time.Second)
		}
	}()
}

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		battles, _ = strconv.ParseFloat(string(msg), bitSize)
	}
}

func getBattles() {
	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	socketUrl := "ws://" + os.Getenv("GAME_SERVER_IP") + ":" + os.Getenv("GAME_SERVER_PORT") + "/" + os.Getenv("WS_PATH")
	fmt.Println("Open connection to " + socketUrl)
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case <-time.After(time.Duration(1) * time.Second * 30):
			// Send an echo packet every second
			err := conn.WriteMessage(websocket.TextMessage, []byte(os.Getenv("GAME_SERVER_TOKEN")))
			if err != nil {
				log.Println("Error during writing to websocket:", err)
				return
			}

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}

func getHostName() string {
	name, err := os.Hostname()
	if err != nil {
		panic("Can't get hostname")
	}
	return name
}

func main() {
	// register metrics
	prometheus.MustRegister(hostName)

	prometheus.MustRegister(battlesNumber)

	// start to count battle
	go getBattles()

	// record metrics
	recordMetrics()

	// start serving http metrics
	fmt.Println("Start metrics at :9101/metrics")
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9101", nil))
}
