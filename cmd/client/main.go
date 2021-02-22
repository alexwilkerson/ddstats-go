package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexwilkerson/ddstats-go/pkg/api"
	"github.com/alexwilkerson/ddstats-go/pkg/config"
	"github.com/alexwilkerson/ddstats-go/pkg/devildaggers"
	"github.com/alexwilkerson/ddstats-go/pkg/socketio"
)

const (
	// version must be in "X.X.X" order.
	version        = "0.6.0"
	consoleTitle   = "ddstats v" + version
	v3survivalHash = "569fead87abf4d30fdee4231a6398051"
)

func main() {
	apiClient, err := api.New("https://ddstats.com")
	if err != nil {
		log.Fatal(err)
	}

	result, err := apiClient.InitConnection(version)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", result)

	gameID, err := apiClient.SubmitGame(
		&api.SubmitGameInput{
			PlayerID:     151675,
			PlayerName:   "VHS",
			Granularity:  1,
			Timer:        0.01,
			TimerSlice:   []float32{0, 0.05, 0.01},
			TotalGems:    0,
			Version:      version,
			SurvivalHash: v3survivalHash,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(gameID)
}

func main4() {
	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", config)
}

func main3() {
	sioClient, err := socketio.New("https://ddstats.com")
	if err != nil {
		log.Fatal(err)
	}

	err = sioClient.Connect(151675)
	if err != nil {
		log.Fatal(err)
	}
	defer sioClient.Disconnect()

	fmt.Scanln()
}

func main2() {
	dd := devildaggers.New()

	connected, err := dd.Connect()
	if err != nil {
		fmt.Println("connection unsuccessful")
		fmt.Println(err)
		os.Exit(1)
	}

	if connected {
		fmt.Println("connection successful")
	}

	err = dd.RefreshData()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(dd.GetLevelHashMD5())
	fmt.Println(len(dd.GetReplayPlayerName()))
	fmt.Println(dd.GetReplayPlayerName())
}
