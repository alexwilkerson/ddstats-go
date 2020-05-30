//go:generate goversioninfo -icon=icon.ico
// +build 386

package main

// wrike integration test

import (
	"github.com/BurntSushi/toml"
	ui "github.com/gizak/termui"
)

func main() {
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
