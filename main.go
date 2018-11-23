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

	fmt.Println(exeFilePath)
	fmt.Println(survivalFilePath)

	fmt.Println()

	fmt.Printf("homing: %v (%T)\n", homing.GetVariable(), homing.GetVariable())
	fmt.Printf("gems: %v (%T)\n", gems.GetVariable(), gems.GetVariable())

	fmt.Println()

	fmt.Printf("timer: %v (%T)\n", timer.GetVariable(), timer.GetVariable())
	fmt.Printf("playerID: %v (%T)\n", playerID.GetVariable(), playerID.GetVariable())
	fmt.Printf("playerName: %v (%T)\n", playerName.GetVariable(), playerName.GetVariable())
	fmt.Printf("playerName length: %v (%T)\n", len(playerName.GetVariable().(string)), len(playerName.GetVariable().(string)))
	fmt.Printf("replayPlayerID: %v (%T)\n", replayPlayerID.GetVariable(), replayPlayerID.GetVariable())
	fmt.Printf("replayPlayerName: %v (%T)\n", replayPlayerName.GetVariable(), replayPlayerName.GetVariable())
	fmt.Printf("totalGems: %v (%T)\n", totalGems.GetVariable(), totalGems.GetVariable())
	fmt.Printf("daggersFired: %v (%T)\n", daggersFired.GetVariable(), daggersFired.GetVariable())
	fmt.Printf("daggersHit: %v (%T)\n", daggersHit.GetVariable(), daggersHit.GetVariable())
	fmt.Printf("enemiesAlive: %v (%T)\n", enemiesAlive.GetVariable(), enemiesAlive.GetVariable())
	fmt.Printf("enemiesKilled: %v (%T)\n", enemiesKilled.GetVariable(), enemiesKilled.GetVariable())
	fmt.Printf("deathType: %v (%T)\n", deathType.GetVariable(), deathType.GetVariable())
	fmt.Printf("isAlive: %v (%T)\n", isAlive.GetVariable(), isAlive.GetVariable())
	fmt.Printf("isReplay: %v (%T)\n", isReplay.GetVariable(), isReplay.GetVariable())
	// fmt.Scanln()
}
