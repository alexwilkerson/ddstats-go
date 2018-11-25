package main

import (
	ui "github.com/gizak/termui"
)

func setupHandles() {
	ui.Handle("<f10>", quit)
}

func quit(ui.Event) {
	ui.StopLoop()
}
