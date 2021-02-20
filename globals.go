package main

import (
	"github.com/TheTitanrain/w32"
)

const (
	version        = "0.6.0"
	v3survivalHash = "569fead87abf4d30fdee4231a6398051"
	captureFPS     = 36
	sioFPS         = 3
	uiFPS          = 2
)

const ddDeathStatus = int32(7)

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
	ready              = true
	debug              = debugLog{log: "[ddstats]\n\n"}
	debugWindowVisible = false
	motd               string
	configReadError    = false
	validVersion       = true
	updateAvailable    bool
	survivalHash       string
	lastGameURL        = "None."
)

var deathTypes = [...]string{"Fallen", "Swarmed", "Impaled", "Gored", "Infested", "Opened", "Purged",
	"Desecrated", "Sacrificed", "Eviscerated", "Annihilated", "Intoxicated",
	"Envenmonated", "Incarnated", "Discarnated", "Barbed"}

var (
	handle           w32.HANDLE
	exeBaseAddress   address
	exeFilePath      string
	survivalFilePath string
	attached         bool
	gameCapture      GameCapture
	gameRecording    GameRecording
	sioVariables     SioVariables
	statDisplay      StatDisplay
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
	timer         = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1A0}, variable: float32(0.0)}
	playerID      = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x5C}, variable: int32(0)}
	totalGems     = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1C0}, variable: int32(0)}
	daggersFired  = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1B4}, variable: int32(0)}
	daggersHit    = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1B8}, variable: int32(0)}
	enemiesAlive  = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1FC}, variable: int32(0)}
	enemiesKilled = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1BC}, variable: int32(0)}
	deathType     = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1C4}, variable: int32(0)}
	isAlive       = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x1A4}, variable: false}
	isReplay      = gameVariable{parentOffset: gameStatsAddress, offsets: []address{0x35D}, variable: false}
)

var (
	homing = gameVariable{parentOffset: gameAddress, offsets: []address{0x0, 0x224}, variable: int32(0)}
	gems   = gameVariable{parentOffset: gameAddress, offsets: []address{0x0, 0x218}, variable: int32(0)}
	isDead = gameVariable{parentOffset: gameAddress, offsets: []address{0x0, 0xCC}, variable: int32(0)}
)

var replayPlayerID = gameReplayIDVariable{
	replayIDVariable: gameVariable{parentOffset: 0x001F80B0, offsets: []address{0x0, 0x18, 0xC, 0x4642}, variable: "XXXXXX"},
}
