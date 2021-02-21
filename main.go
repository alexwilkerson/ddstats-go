package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
	ui "github.com/gizak/termui"
)

func main() {
	fmt.Println("ran")
	getHandle()
	if handle == 0 {
		fmt.Println("nope")
	}
	fmt.Println("got handle")
	fmt.Printf("player id: %d\n", gameCapture.playerID)
	gameCapture.GetPlayerVariables()
	fmt.Printf("player id: %d\n", gameCapture.playerID)
}

func main2() {
	setConsoleTitle("ddstats v" + version)

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		configReadError = true
		ready = false
	}

	debug.Log(config.Stream.Stats)

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	if !config.OfflineMode {
		go getMotd()
		if config.Stream.Stats || config.Stream.ReplayStats || config.Stream.NonDefaultSpawnsets {
			go liveStreamStats()
		}
	}

	go gameCapture.Run()

	go classicLayout()

	uiLoop()
}
