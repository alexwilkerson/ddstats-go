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
	replayPlayerName.Get()
	totalGems.Get()
	daggersFired.Get()
	daggersHit.Get()
	enemiesAlive.Get()
	enemiesKilled.Get()
	deathType.Get()
	isAlive.Get()
	isReplay.Get()

	replayPlayerID.Get()

	fmt.Printf("homing: %v\n", homing.variable)
	fmt.Printf("gems: %v\n", gems.variable)
	fmt.Println()
	fmt.Printf("timer: %v\n", timer.variable)
	fmt.Printf("playerID: %v\n", playerID.variable)
	fmt.Printf("playerName: %v\n", playerName.variable)
	fmt.Printf("replayPlayerID: %v\n", replayPlayerID.variable)
	fmt.Printf("replayPlayerName: %v\n", replayPlayerName.variable)
	fmt.Printf("totalGems: %v\n", totalGems.variable)
	fmt.Printf("daggersFired: %v\n", daggersFired.variable)
	fmt.Printf("daggersHit: %v\n", daggersHit.variable)
	fmt.Printf("enemiesAlive: %v\n", enemiesAlive.variable)
	fmt.Printf("enemiesKilled: %v\n", enemiesKilled.variable)
	fmt.Printf("deathType: %v\n", deathType.variable)
	fmt.Printf("isAlive: %v\n", isAlive.variable)
	fmt.Printf("isReplay: %v\n", isReplay.variable)
}
