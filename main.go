package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type InternalClaims struct {
	ClientID string `json:"client_id,omitempty"`
	Role     string `json:"role,omitempty"`
	jwt.StandardClaims
}
type Data struct {
	TotalBattle        int `json:"total_battle"`
	TotalPlayer        int `json:"total_player"`
	TotalConnection    int `json:"total_connection"`
	ClientMap          int `json:"client_map"`
	MmrDataQueue       int `json:"mmr_data_queue"`
	DivisionDataQueue  int `json:"division_data_queue"`
	MmrReadyQueue      int `json:"mmr_ready_queue"`
	DivisionReadyQueue int `json:"division_ready_queue"`
	PveQueue           int `json:"pve_queue"`
	TotalPve           int `json:"total_pve"`
	TotalPvp           int `json:"total_pvp"`

	PveQueueCap           int `json:"pve_queue_cap"`
	MmrDataQueueCap       int `json:"mmr_data_queue_cap"`
	DivisionDataQueueCap  int `json:"division_data_queue_cap"`
	MmrReadyQueueCap      int `json:"mmr_ready_queue_cap"`
	DivisionReadyQueueCap int `json:"division_ready_queue_cap"`
}

var battlesValue, clientmapValue, mmrdataqueueValue, divisiondataqueueValue, mmrreadyqueueValue, divisionreadyqueueValue, pvequeueValue, playerValue, connectionValue, pveValue, pvpValue, PveQueueCapValue, MmrDataQueueCapValue, DivisionDataQueueCapValue, MmrReadyQueueCapValue, DivisionReadyQueueCapValue int

var (
	PveQueueCapNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_pve_queue_cap",
		Help:        "The total number of pve hashmap",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})
	MmrDataQueueCapNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_mmr_data_queue_cap",
		Help:        "The total number of pvp hashmap",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	DivisionDataQueueCapNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_division_data_queue_cap",
		Help:        "The total number of division data queue",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})
	MmrReadyQueueCapNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_serve_mmr_ready_queue_cap",
		Help:        "The total number of mmr queue cap",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	DivisionReadyQueueCapNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_division_ready_queue_cap",
		Help:        "The total number of pvp hashmap",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	battlesNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_battle_total",
		Help:        "The total number of match happening",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	pveNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_pve_total",
		Help:        "The total number of PvE happening",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	pvpNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_pvp_total",
		Help:        "The total number of PvP happening",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	playerNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_total_player",
		Help:        "The total player.",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	connectionNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_total_connection",
		Help:        "The total connections.",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	clientMap = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_client_map",
		Help:        "The total client map",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	mmrDataQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_mmr_data_queue",
		Help:        "The total MMR data in queue.",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	divisionDataQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_division_data_queue",
		Help:        "The total division data in queue.",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	mmrReadyQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_mmr_ready_queue",
		Help:        "The total MMR ready in queue.",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	divisionReadyQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_division_ready_queue",
		Help:        "The total division ready in queue.",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})

	pveQueue = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "origin_game_server_pve_queue",
		Help:        "The total PvE in queue.",
		ConstLabels: map[string]string{"nodename": getHostName()},
	})
)

// Gen token
func GenerateToken() (string, error) {
	claims := &InternalClaims{
		ClientID: "internal",
		Role:     "internal",
		StandardClaims: jwt.StandardClaims{
			Issuer:    "AxieInfinity",
			IssuedAt:  time.Now().UTC().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("GAME_SERVER_JWT")))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// set metrics
func recordMetrics() {
	go func() {
		for {
			// get battles
			getBattles()
			// set metrics
			battlesNumber.Set(float64(battlesValue))
			clientMap.Set(float64(clientmapValue))
			mmrDataQueue.Set(float64(mmrdataqueueValue))
			divisionDataQueue.Set(float64(divisiondataqueueValue))
			mmrReadyQueue.Set(float64(mmrreadyqueueValue))
			divisionReadyQueue.Set(float64(divisionreadyqueueValue))
			pveQueue.Set(float64(pvequeueValue))
			playerNumber.Set(float64(playerValue))
			connectionNumber.Set(float64(connectionValue))
			pvpNumber.Set((float64(pvpValue)))
			pveNumber.Set((float64(pveValue)))
			PveQueueCapNumber.Set((float64(PveQueueCapValue)))
			MmrDataQueueCapNumber.Set(float64(MmrDataQueueCapValue))
			DivisionDataQueueCapNumber.Set(float64(DivisionDataQueueCapValue))
			MmrReadyQueueCapNumber.Set(float64(MmrReadyQueueCapValue))
			DivisionReadyQueueCapNumber.Set(float64(DivisionReadyQueueCapValue))
			time.Sleep(15 * time.Second)
		}
	}()
}

func getBattles() {
	// generate token
	token, err := GenerateToken()
	if err != nil {
		panic(err)
	}

	// game server metrics path
	httpUrl := "http://" + os.Getenv("GAME_SERVER_IP") + ":" + os.Getenv("GAME_SERVER_PORT") + "/" + os.Getenv("BATTLE_PATH")
	fmt.Println("Open connection to " + httpUrl)

	req, err := http.NewRequest("GET", httpUrl, nil)
	if err != nil {
		panic(err)
	}

	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{token},
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
	fmt.Println(string(body))
	dat := Data{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}

	// set value
	battlesValue = dat.TotalBattle
	clientmapValue = dat.ClientMap
	mmrdataqueueValue = dat.MmrDataQueue
	divisiondataqueueValue = dat.DivisionDataQueue
	mmrreadyqueueValue = dat.MmrReadyQueue
	divisionreadyqueueValue = dat.DivisionReadyQueue
	pvequeueValue = dat.PveQueue
	playerValue = dat.TotalPlayer
	connectionValue = dat.TotalConnection
	pveValue = dat.TotalPve
	pvpValue = dat.TotalPvp
	PveQueueCapValue = dat.PveQueueCap
	MmrDataQueueCapValue = dat.MmrDataQueueCap
	DivisionDataQueueCapValue = dat.DivisionDataQueueCap
	MmrReadyQueueCapValue = dat.MmrReadyQueueCap
	DivisionReadyQueueCapValue = dat.DivisionReadyQueueCap
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
	prometheus.MustRegister(clientMap)
	prometheus.MustRegister(mmrDataQueue)
	prometheus.MustRegister(divisionDataQueue)
	prometheus.MustRegister(mmrReadyQueue)
	prometheus.MustRegister(divisionReadyQueue)
	prometheus.MustRegister(pveQueue)
	prometheus.MustRegister(connectionNumber)
	prometheus.MustRegister(playerNumber)
	prometheus.MustRegister(pveNumber)
	prometheus.MustRegister(pvpNumber)
	prometheus.MustRegister(PveQueueCapNumber)
	prometheus.MustRegister(MmrDataQueueCapNumber)
	prometheus.MustRegister(DivisionDataQueueCapNumber)
	prometheus.MustRegister(MmrReadyQueueCapNumber)
	prometheus.MustRegister(DivisionReadyQueueCapNumber)
	// record metrics
	recordMetrics()

	// start serving http metrics
	fmt.Println("Start metrics at :9101/metrics")
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9101", nil))
}
