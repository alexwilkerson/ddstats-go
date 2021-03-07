//go:generate goversioninfo

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexwilkerson/ddstats-go/pkg/client"
)

const (
	// version must be in "X.X.X" order.
	version        = "0.6.2"
	consoleTitle   = "ddstats v" + version
	v3survivalHash = "569fead87abf4d30fdee4231a6398051"
	grpcAddr       = "172.104.11.117:80"
)

func main() {
	client, err := client.New(version, grpcAddr)
	if err != nil {
		if err := logError(err); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}
	err = client.Run()
	if err != nil {
		if err := logError(err); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}
}

func logError(inputErr error) error {
	f, err := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("logError: error opening file: %w", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Printf("%v\n", inputErr)
	return nil
}

// // func main6() {
// // 	ui, err := consoleui.New()
// // 	if err != nil {
// // 		log.Fatal(err)
// // 	}
// // 	defer ui.Close()

// // 	data := consoleui.Data{
// // 		PlayerName:      "VHS",
// // 		Version:         version,
// // 		UpdateAvailable: true,
// // 		MOTD:            "hello there",
// // 	}

// // 	ui.DrawScreen(&data)

// // 	fmt.Scanln()
// // }

// func main5() {
// 	apiClient, err := api.New("https://ddstats.com")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	result, err := apiClient.InitConnection(version)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("%v\n", result)

// 	gameID, err := apiClient.SubmitGame(
// 		&api.SubmitGameInput{
// 			PlayerID:     151675,
// 			PlayerName:   "VHS",
// 			Granularity:  1,
// 			Timer:        0.01,
// 			TimerSlice:   []float32{0, 0.05, 0.01},
// 			TotalGems:    0,
// 			Version:      version,
// 			SurvivalHash: v3survivalHash,
// 		},
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println(gameID)
// }

// func main4() {
// 	config, err := config.New()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("%v\n", config)
// }

// func main3() {
// 	sioClient, err := socketio.New("https://ddstats.com")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	err = sioClient.Connect(151675)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer sioClient.Disconnect()

// 	fmt.Scanln()
// }

// func main() {
// 	dd := devildaggers.New()

// 	connected, err := dd.Connect()
// 	if err != nil {
// 		fmt.Println("connection unsuccessful")
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	if connected {
// 		fmt.Println("connection successful")
// 	}

// 	err = dd.RefreshData()
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	fmt.Println(dd.GetDDStatsVersion())
// 	fmt.Println(dd.GetPlayerID())
// 	fmt.Println(dd.GetPlayerName())
// }
