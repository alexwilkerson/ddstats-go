package main

import (
	"fmt"
	"log"

	"github.com/TheTitanrain/w32"
)

func main() {
	var err error
	handle, err = getHandle()
	if err != nil {
		log.Fatal(err)
	}
	defer w32.CloseHandle(handle)

	homing.Get()
	gems.Get()

	timer.Get()
	playerID.Get()
	playerName.Get()
	totalGems.Get()
	daggersFired.Get()
	daggersHit.Get()
	enemiesAlive.Get()
	enemiesKilled.Get()
	deathType.Get()
	isAlive.Get()
	isReplay.Get()

	if isReplay.GetVariable().(bool) {
		replayPlayerName.Get()
		replayPlayerID.Get()
	}

	fmt.Printf("homing: %v\n", homing.GetVariable())
	fmt.Printf("gems: %v\n", gems.GetVariable())
	fmt.Println()
	fmt.Printf("timer: %v\n", timer.GetVariable())
	fmt.Printf("playerID: %v\n", playerID.GetVariable())
	fmt.Printf("playerName: %v\n", playerName.GetVariable())
	fmt.Printf("replayPlayerID: %v\n", replayPlayerID.GetVariable())
	fmt.Printf("replayPlayerName: %v\n", replayPlayerName.GetVariable())
	fmt.Printf("totalGems: %v\n", totalGems.GetVariable())
	fmt.Printf("daggersFired: %v\n", daggersFired.GetVariable())
	fmt.Printf("daggersHit: %v\n", daggersHit.GetVariable())
	fmt.Printf("enemiesAlive: %v\n", enemiesAlive.GetVariable())
	fmt.Printf("enemiesKilled: %v\n", enemiesKilled.GetVariable())
	fmt.Printf("deathType: %v\n", deathType.GetVariable())
	fmt.Printf("isAlive: %v\n", isAlive.GetVariable())
	fmt.Printf("isReplay: %v\n", isReplay.GetVariable())

	fmt.Scanln()
}
