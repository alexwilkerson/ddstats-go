package main

import (
	"github.com/TheTitanrain/w32"
)

const (
	version    = "0.4.0"
	captureFPS = 60
)

const (
	gameStatsAddress address = 0x001F30C0
	gameAddress      address = 0x001F8084
)

type status int

const (
	statusNotConnected status = iota
	statusConnecting
	statusIsPlaying
	statusIsReplay
	statusInMainMenu
	statusInDaggerLobby
	statusIsDead
)

var (
	motd            string
	validVersion    = true
	updateAvailable bool
)

var (
	handle           w32.HANDLE
	exeBaseAddress   address
	exeFilePath      string
	survivalFilePath string
	attached         bool
	gc               gameCapture
	sd               statDisplay
)

var (
	playerName = gameStringVariable{
		stringVariable: gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x60}, variable: ""},
	}
	replayPlayerName = gameStringVariable{
		stringVariable: gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x360}, variable: ""},
	}
)

var (
	level2time float32
	level3time float32
	level4time float32
)

var (
	timer         = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1A0}, variable: 0.0}
	playerID      = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x5C}, variable: 0}
	totalGems     = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1C0}, variable: 0}
	daggersFired  = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1B4}, variable: 0}
	daggersHit    = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1B8}, variable: 0}
	enemiesAlive  = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1FC}, variable: 0}
	enemiesKilled = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1BC}, variable: 0}
	deathType     = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1C4}, variable: 0}
	isAlive       = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1A4}, variable: false}
	isReplay      = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x35D}, variable: false}
)

var (
	homing = gameVariable{parentOffset: gameAddress, offsets: []address{0x0, 0x224}, variable: 0}
	gems   = gameVariable{parentOffset: gameAddress, offsets: []address{0x0, 0x218}, variable: 0}
)

var replayPlayerID = gameReplayIDVariable{
	replayIDVariable: gameVariable{parentOffset: 0x001F80B0, offsets: []address{0x0, 0x18, 0xC, 0x4642}, variable: "XXXXXX"},
}
