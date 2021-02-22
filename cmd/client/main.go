package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexwilkerson/ddstats-go/pkg/config"
	"github.com/alexwilkerson/ddstats-go/pkg/devildaggers"
	"github.com/alexwilkerson/ddstats-go/pkg/socketio"
)

const (
	consoleTitle = "ddstats v0.6.0"
)

func main() {
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
