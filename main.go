//go:generate goversioninfo -icon=icon.ico
package main

import (
	ui "github.com/gizak/termui"
)

func main() {
	setConsoleTitle("ddstats v" + version)

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	setupHandles()

	go getMotd()

	go liveStreamStats()

	go gameCapture.Run()

	go classicLayout()

	ui.Loop()
}
