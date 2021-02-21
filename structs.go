package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"
)

type debugLog struct {
	t   time.Time
	log string
}

func (dl *debugLog) Log(i interface{}) {
	dl.t = time.Now()
	dl.log = fmt.Sprintf("%02d:%02d:%02d: ", dl.t.Hour(), dl.t.Minute(), dl.t.Second()) + fmt.Sprintf("%v", i) + "\n" + dl.log
}

func (dl *debugLog) Clear() {
	dl.log = ""
}

type address uintptr

type gameVariable struct {
	parentOffset     address
	offsets          []address
	previousVariable interface{}
	variable         interface{}
}

func (gv *gameVariable) Get() {
	gv.previousVariable = gv.variable
	var pointer address
	getValue(&pointer, exeBaseAddress+gv.parentOffset)
	for i := 0; i < len(gv.offsets)-1; i++ {
		getValue(&pointer, pointer+gv.offsets[i])
	}
	getValue(&gv.variable, pointer+gv.offsets[len(gv.offsets)-1])
	switch gv.variable.(type) {
	case int:
		gv.variable = int32(gv.variable.(int))
	case int32:
		gv.variable = gv.variable.(int32)
	case float32:
		gv.variable = float32(gv.variable.(float32))
	case float64:
		gv.variable = float64(gv.variable.(float64))
	case bool:
		gv.variable = bool(gv.variable.(bool))
	case string:
		gv.variable = string(gv.variable.(string))
	case address:
		gv.variable = address(gv.variable.(address))
	case uintptr:
		gv.variable = uintptr(gv.variable.(uintptr))
	}
}

func (gv *gameVariable) AsyncGet(wg *sync.WaitGroup) {
	gv.Get()
	wg.Done()
}

func (gv *gameVariable) Reset(i interface{}) {
	gv.variable = i
	gv.previousVariable = i
}

func (gv *gameVariable) GetPreviousVariable() interface{} {
	return gv.previousVariable
}

func (gv *gameVariable) GetVariable() interface{} {
	return gv.variable
}

type gameStringVariable struct {
	stringVariable gameVariable
	variable       string
}

// maxSize is used to check the maximum size of the string array in dd.
// if maxSize is 15, the pointer points directly to the char array
// if maxSize is 31 the char array holds the address of where the new
// string is stored. the offset of the maxSize is always 0x14.
func (gsv *gameStringVariable) Get() {
	lengthOffset := gsv.stringVariable.offsets[0] + 0x10
	lengthVariable := gameVariable{parentOffset: gameStatsAddress, offsets: []address{lengthOffset}, variable: 0}
	lengthVariable.Get()
	length := lengthVariable.variable

	maxSizeOffset := gsv.stringVariable.offsets[0] + 0x14
	maxSizeVariable := gameVariable{parentOffset: gameStatsAddress, offsets: []address{maxSizeOffset}, variable: 0}
	maxSizeVariable.Get()
	maxSize := maxSizeVariable.variable.(int32)

	iterations := int(math.Log2(float64(maxSize)+1)) - 4

	for i := 0; i < iterations; i++ {
		gsv.stringVariable.offsets = append(gsv.stringVariable.offsets, 0x0)
	}
	gsv.stringVariable.variable = string(make([]byte, length.(int32)))
	gsv.stringVariable.Get()

	gsv.variable = gsv.stringVariable.variable.(string)[:length.(int32)]
	gsv.stringVariable.variable = ""
}

func (gsv *gameStringVariable) GetVariable() interface{} {
	return string(gsv.variable)
}

type gameReplayIDVariable struct {
	replayIDVariable gameVariable
	variable         int32
}

func (gridv *gameReplayIDVariable) GetVariable() interface{} {
	return int(gridv.variable)
}

func (gridv *gameReplayIDVariable) Get() {
	gridv.replayIDVariable.Get()
	gridv.variable = gridv.replayIDVariable.variable.(int32)
	gridv.replayIDVariable.variable = "XXXXXX"
}

type GameRecording struct {
	PlayerID            int32     `json:"playerID"`
	PlayerName          string    `json:"playerName"`
	Granularity         int32     `json:"granularity"`
	Timer               float32   `json:"inGameTimer"`
	TimerSlice          []float32 `json:"inGameTimerVector"`
	TotalGems           int32     `json:"gems"`
	TotalGemsSlice      []int32   `json:"gemsVector"`
	Level2time          float32   `json:"levelTwoTime"`
	Level3time          float32   `json:"levelThreeTime"`
	Level4time          float32   `json:"levelFourTime"`
	Homing              int32     `json:"homingDaggers"`
	HomingSlice         []int32   `json:"homingDaggersVector"`
	HomingMax           int32     `json:"homingDaggersMax"`
	HomingMaxTime       float32   `json:"homingDaggersMaxTime"`
	DaggersFired        int32     `json:"daggersFired"`
	DaggersFiredSlice   []int32   `json:"daggersFiredVector"`
	DaggersHit          int32     `json:"daggersHit"`
	DaggersHitSlice     []int32   `json:"daggersHitVector"`
	EnemiesAlive        int32     `json:"enemiesAlive"`
	EnemiesAliveSlice   []int32   `json:"enemiesAliveVector"`
	EnemiesAliveMax     int32     `json:"enemiesAliveMax"`
	EnemiesAliveMaxTime float32   `json:"enemiesAliveMaxTime"`
	EnemiesKilled       int32     `json:"enemiesKilled"`
	EnemiesKilledSlice  []int32   `json:"enemiesKilledVector"`
	DeathType           int32     `json:"deathType"`
	ReplayPlayerID      int32     `json:"replayPlayerID"`
	Version             string    `json:"version"`
	SurvivalHash        string    `json:"survivalHash"`
}

func (gr *GameRecording) RecordVariables() {
	gr.TimerSlice = append(gr.TimerSlice, gameCapture.timer)
	gr.TotalGemsSlice = append(gr.TotalGemsSlice, gameCapture.totalGems)
	gr.HomingSlice = append(gr.HomingSlice, gameCapture.homing)
	gr.DaggersFiredSlice = append(gr.DaggersFiredSlice, gameCapture.daggersFired)
	gr.DaggersHitSlice = append(gr.DaggersHitSlice, gameCapture.daggersHit)
	gr.EnemiesAliveSlice = append(gr.EnemiesAliveSlice, gameCapture.enemiesAlive)
	gr.EnemiesKilledSlice = append(gr.EnemiesKilledSlice, gameCapture.enemiesKilled)
}

func (gr *GameRecording) Stop() {
	// gameCapture.GetPlayerVariables()
	gr.PlayerID = gameCapture.playerID
	gr.PlayerName = gameCapture.playerName
	gr.Timer = gameCapture.timer
	gr.TimerSlice = append(gr.TimerSlice, gameCapture.timer)
	gr.TotalGems = gameCapture.totalGemsAtDeath
	gr.TotalGemsSlice = append(gr.TotalGemsSlice, gameCapture.totalGemsAtDeath)
	gr.Level2time = gameCapture.level2time
	gr.Level3time = gameCapture.level3time
	gr.Level4time = gameCapture.level4time
	gr.Homing = gameCapture.homing
	gr.HomingSlice = append(gr.HomingSlice, gameCapture.homing)
	gr.HomingMax = gameCapture.homingMax
	gr.HomingMaxTime = gameCapture.homingMaxTime
	gr.DaggersFired = gameCapture.daggersFired
	gr.DaggersFiredSlice = append(gr.DaggersFiredSlice, gameCapture.daggersFired)
	gr.DaggersHit = gameCapture.daggersHit
	gr.DaggersHitSlice = append(gr.DaggersHitSlice, gameCapture.daggersHit)
	gr.EnemiesAlive = gameCapture.enemiesAlive
	gr.EnemiesAliveSlice = append(gr.EnemiesAliveSlice, gameCapture.enemiesAlive)
	gr.EnemiesAliveMax = gameCapture.enemiesAliveMax
	gr.EnemiesAliveMaxTime = gameCapture.enemiesAliveMaxTime
	gr.EnemiesKilled = gameCapture.enemiesKilled
	gr.EnemiesKilledSlice = append(gr.EnemiesKilledSlice, gameCapture.enemiesKilled)
	gr.DeathType = gameCapture.deathType
	gr.ReplayPlayerID = gameCapture.replayPlayerID
	if gr.SurvivalHash == "" {
		gameCapture.GetSurvivalHash()
		gr.SurvivalHash = gameCapture.survivalHash
	}
}

func (gr *GameRecording) Reset() {
	gr.PlayerID = gameCapture.playerID
	gr.PlayerName = gameCapture.playerName
	gr.Granularity = 1
	gr.Timer = -1.0
	gr.TimerSlice = []float32{}
	gr.TotalGems = 0
	gr.TotalGemsSlice = []int32{}
	gr.Level2time = 0
	gr.Level3time = 0
	gr.Level4time = 0
	gr.Homing = 0
	gr.HomingSlice = []int32{}
	gr.HomingMax = 0
	gr.HomingMaxTime = 0.0
	gr.DaggersFired = 0
	gr.DaggersFiredSlice = []int32{}
	gr.DaggersHit = 0
	gr.DaggersHitSlice = []int32{}
	gr.EnemiesAlive = 0
	gr.EnemiesAliveSlice = []int32{}
	gr.EnemiesAliveMax = 0
	gr.EnemiesAliveMaxTime = 0.0
	gr.EnemiesKilled = 0
	gr.EnemiesKilledSlice = []int32{}
	gr.DeathType = 0
	gr.ReplayPlayerID = 0
	gr.Version = version
	gr.SurvivalHash = gameCapture.survivalHash
}

func (gr *GameRecording) WasReset() bool {
	if len(gr.TimerSlice) == 0 {
		return true
	}
	return false
}

// GameCapture is the structure that captures all game data from dd.exe.
type GameCapture struct {
	wg                       sync.WaitGroup
	status                   status
	lastRecording            float32
	survivalHash             string
	v3                       bool
	isAlive                  bool
	isReplay                 bool
	playerID                 int32
	playerName               string
	deathType                int32
	replayPlayerID           int32
	replayPlayerName         string
	timer                    float32
	gems                     int32
	totalGems                int32
	totalGemsAtDeath         int32
	level2time               float32
	level3time               float32
	level4time               float32
	homing                   int32
	homingMax                int32
	homingMaxPerSecond       int32
	homingMaxTime            float32
	enemiesAlive             int32
	enemiesAliveMaxPerSecond int32
	enemiesAliveMax          int32
	enemiesAliveMaxTime      float32
	enemiesKilled            int32
	daggersFired             int32
	daggersHit               int32
	accuracy                 float32
}

func (gc *GameCapture) Reset() {
	gc.lastRecording = 0.0
	gc.deathType = 0
	deathType.Reset(0)
	gc.timer = float32(0.0)
	timer.Reset(0.0)
	gc.gems = 0
	gems.Reset(0)
	gc.totalGems = 0
	gc.totalGemsAtDeath = 0
	totalGems.Reset(0)
	gc.level2time = 0.0
	gc.level3time = 0.0
	gc.level4time = 0.0
	gc.homing = 0
	homing.Reset(0)
	gc.homingMax = 0
	gc.homingMaxPerSecond = 0
	gc.homingMaxTime = 0.0
	gc.enemiesAlive = 0
	gc.enemiesAliveMaxPerSecond = 0
	enemiesAlive.Reset(0)
	gc.enemiesAliveMax = 0
	gc.enemiesAliveMaxTime = 0.0
	gc.enemiesKilled = 0
	enemiesKilled.Reset(0)
	gc.daggersFired = 0
	daggersFired.Reset(0)
	gc.daggersHit = 0
	daggersHit.Reset(0)
	gc.accuracy = 0.0
}

func (gc *GameCapture) GetSurvivalHash() error {
	f, err := os.Open(survivalFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	gc.survivalHash = fmt.Sprintf("%x", h.Sum(nil))
	if gc.survivalHash == v3survivalHash {
		gc.v3 = true
	} else {
		gc.v3 = false
	}
	return nil
}

func (gc *GameCapture) GetPlayerVariables() {
	if handle != 0 {
		playerID.Get()
		// playerName.Get()

		gc.playerID = playerID.GetVariable().(int32)
		// gc.playerName = playerName.GetVariable().(string)
	}
}

func (gc *GameCapture) GetReplayPlayerVariables() {
	if handle != 0 {
		replayPlayerName.Get()
		// This is necessary, because when you watch your own replays, there
		// is no replayPlayerID variable.
		if replayPlayerName.GetVariable() == playerName.GetVariable() {
			gc.replayPlayerID = playerID.GetVariable().(int32)
		} else {
			replayPlayerID.Get()
			gc.replayPlayerID = int32(replayPlayerID.GetVariable().(int))
		}
		gc.replayPlayerName = replayPlayerName.GetVariable().(string)
	}
}

func (gc *GameCapture) ResetReplayPlayerVariables() {
	gc.replayPlayerID = 0
	gc.replayPlayerName = ""
}

func (gc *GameCapture) GetStatus() status {
	return gc.status
}

// This function is heavily commented because of the weird nature
// of reading how another program's memory is working. The logic
// is a bit squirrely.
func (gc *GameCapture) GetGameVariables() {

	// Get the handle before retrieving any variables. This helps ensure that the
	// user hasn't closed dd in the interim.
	getHandle()

	// If the game is not found, reset the playerName to ""
	// This is to make sure that 'connecting' is appropriately set
	// on reopening dd.exe multiple times.
	if handle == 0 {

		gc.status = statusNotConnected
		gc.playerName = ""
		gc.playerID = 0

	} else {

		// if you have a handle, grab all of the variables from the game
		gc.wg.Add(11)
		go isDead.AsyncGet(&gc.wg)
		go isAlive.AsyncGet(&gc.wg)
		go isReplay.AsyncGet(&gc.wg)
		go timer.AsyncGet(&gc.wg)
		go gems.AsyncGet(&gc.wg)
		go totalGems.AsyncGet(&gc.wg)
		go homing.AsyncGet(&gc.wg)
		go enemiesAlive.AsyncGet(&gc.wg)
		go enemiesKilled.AsyncGet(&gc.wg)
		go daggersFired.AsyncGet(&gc.wg)
		go daggersHit.AsyncGet(&gc.wg)
		gc.wg.Wait()

		gc.isAlive = isAlive.GetVariable().(bool)

		// If dead and previously were alive, this means the player just died
		// so the game must then be completed and sent to the server.
		if isDead.GetPreviousVariable().(int32) != ddDeathStatus && isDead.GetVariable().(int32) == ddDeathStatus && !gameRecording.WasReset() {

			if gc.isReplay {
				gc.GetReplayPlayerVariables()
			}

			gc.status = statusIsDead

			deathType.Get()
			gc.deathType = deathType.GetVariable().(int32)
			// if deathType is invalid, make it 0 (FALLEN) by default. This is just to make sure
			// some strange memory address was read.
			if gc.deathType < 0 || gc.deathType > 15 {
				gc.deathType = 0
			}

			gc.timer = timer.GetVariable().(float32)
			// this accounts for totalGems being reset on death.
			gc.totalGems = totalGems.GetPreviousVariable().(int32)
			// this accounts for homing being reset on death.
			// there might be a more elegent solution for this.
			gc.homing = homing.GetPreviousVariable().(int32)
			if gc.homing > gc.homingMax {
				gc.homingMax = gc.homing
				gc.homingMaxTime = gc.timer
			}
			if gc.homingMaxPerSecond > gc.homing {
				gc.homing = gc.homingMaxPerSecond
			}
			// Sometimes homing drops below zero. This fixes recording that.
			if gc.homing < 0 {
				gc.homing = 0
			}
			if gc.enemiesAlive > gc.enemiesAliveMax {
				gc.enemiesAliveMax = gc.enemiesAlive
				gc.enemiesAliveMaxTime = gc.timer
			}
			// This might be off, may need to do more testing.
			if gc.enemiesAliveMaxPerSecond > gc.enemiesAlive {
				gc.enemiesAlive = gc.enemiesAliveMaxPerSecond
			}
			gc.enemiesAlive = enemiesAlive.GetVariable().(int32)
			gc.enemiesKilled = enemiesKilled.GetVariable().(int32)
			gc.daggersFired = daggersFired.GetVariable().(int32)
			gc.daggersHit = daggersHit.GetVariable().(int32)
			gc.totalGemsAtDeath = totalGems.GetPreviousVariable().(int32)
			if gc.level2time == 0.0 && gc.gems >= 10 {
				gc.level2time = gc.timer
			}
			if gc.level3time == 0.0 && gc.gems == 70 {
				gc.level3time = gc.timer
			}
			if gc.level4time == 0.0 && gc.gems == 71 {
				gc.level4time = gc.timer
			}

			// Stop the game from recording, send a copy to the server,
			// reset the gameRecording struct, update the display,
			// reset the gameCapture struct

			// check one last time to make sure these variables are set.
			// This is to fix a bug I noticed where empty names were
			// submitted to the server
			if gc.playerName == "" || gc.playerID == -1 {
				gc.GetPlayerVariables()
			}

			gameRecording.Stop()
			sioVariables.Update()
			go submitGame(gameRecording)
			gameRecording.Reset()
			statDisplay.Update()
			gameCapture.Reset()
		}

		if gc.isAlive {

			// arbitrary number above the max user id number
			if gc.playerName == "" || gc.playerID == -1 {
				// Sleep for half a second to make sure dd has a chance to initialize memory before reading
				// player variables.
				time.Sleep(time.Second / 2)
				gc.GetPlayerVariables()
			}

			gc.isReplay = isReplay.GetVariable().(bool)
			gc.timer = timer.GetVariable().(float32)
			if gc.timer == 0.0 {
				if enemiesAlive.GetVariable().(int32) == 0 {
					if gc.survivalHash == "" {
						gc.GetSurvivalHash()
					}
					gc.status = statusInDaggerLobby
				} else {
					gc.survivalHash = ""
					gc.status = statusInMainMenu
					sioVariables.Update()
				}
				return
			}

			if gc.isReplay {
				gc.status = statusIsReplay
			} else {
				gc.status = statusIsPlaying
			}

			// The rest is only recorded if isAlive or isReplay...

			// If the user presses 'R' to restart mid-game
			if timer.GetVariable().(float32) < timer.GetPreviousVariable().(float32) {
				gameRecording.Reset()
			}

			// If a new game was started,
			if gameRecording.WasReset() {
				gameRecording.Reset()
				// Inititally lastRecording is set to 1 so that the rest of the logic works correctly.
				gc.lastRecording = gc.timer - 1
				// This is done every game in case the user opens ddstats at some weird point.
				gc.GetPlayerVariables()
				if gc.isReplay {
					gc.GetReplayPlayerVariables()
				} else {
					gc.replayPlayerID = 0
					gc.replayPlayerName = ""
				}
			}

			gc.gems = gems.GetVariable().(int32)
			gc.totalGems = totalGems.GetVariable().(int32)
			gc.homing = homing.GetVariable().(int32)
			if gc.homing > gc.homingMax {
				gc.homingMax = gc.homing
				gc.homingMaxTime = gc.timer
			}
			// Sometimes homing drops below zero. This fixes recording that.
			if gc.homing < 0 {
				gc.homing = 0
			}
			if gc.homingMaxPerSecond > gc.homing {
				gc.homing = gc.homingMaxPerSecond
			}
			if gc.enemiesAlive > gc.enemiesAliveMax {
				gc.enemiesAliveMax = gc.enemiesAlive
				gc.enemiesAliveMaxTime = gc.timer
			}
			gc.enemiesAlive = enemiesAlive.GetVariable().(int32)
			if gc.enemiesAlive > gc.enemiesAliveMaxPerSecond {
				gc.enemiesAliveMaxPerSecond = gc.enemiesAlive
			}
			gc.enemiesKilled = enemiesKilled.GetVariable().(int32)
			gc.daggersFired = daggersFired.GetVariable().(int32)
			gc.daggersHit = daggersHit.GetVariable().(int32)
			if gc.daggersFired > 0 {
				gc.accuracy = (float32(gc.daggersHit) / float32(gc.daggersFired)) * 100
			} else {
				gc.accuracy = 0.0
			}
			if gc.level2time == 0.0 && gc.gems >= 10 {
				gc.level2time = gc.timer
			}
			if gc.level3time == 0.0 && gc.gems == 70 {
				gc.level3time = gc.timer
			}
			if gc.level4time == 0.0 && gc.gems == 71 {
				gc.level4time = gc.timer
			}

			sioVariables.Update()

			// If more than a second has elapsed in the game, capture a recording.
			if math.Floor(float64(gc.timer))-math.Floor(float64(gc.lastRecording)) >= 1 {
				// enemiesAliveMaxPerSecond is a variable used to smooth out the enemiesAlive
				// graph. dd.exe will drop enemiesAlive count to 0 occasionally which results
				// in a jagged, uneven graph. Taking the maximum count over a second fixes this issue.
				if gc.enemiesAliveMaxPerSecond > gc.enemiesAlive {
					gc.enemiesAlive = gc.enemiesAliveMaxPerSecond
				}
				if gc.homingMaxPerSecond > gc.homing {
					gc.homing = gc.homingMaxPerSecond
				}
				// Sometimes at the beginning of the game the enemiesAlive variable does not reset properly.
				// this fixes that.
				if gc.timer < 1.0 {
					gc.enemiesAlive = 0
				}
				gameRecording.RecordVariables()
				gc.lastRecording = gc.timer
				// reset homingMaxPerSecond, enemiesAliveMaxPerSecond to 0 every second.
				gc.homingMaxPerSecond = homing.GetVariable().(int32)
				gc.enemiesAliveMaxPerSecond = enemiesAlive.GetVariable().(int32)
			}
		} else if gc.playerName == "" {
			// The isDead variable has one function... If the user starts ddstats
			// while on the death screen, this variable will be 7 and ddstats will
			// be able to differentiate from connecting to dd.exe.
			isDead.Get()
			if isDead.GetVariable().(int32) == ddDeathStatus {
				gc.GetPlayerVariables()
				gc.status = statusIsDead
			} else {
				gc.status = statusConnecting
			}
		} else {
			if gc.replayPlayerName != "" {
				gc.ResetReplayPlayerVariables()
			}
			gc.status = statusIsDead
			deathType.Get()
			gc.deathType = deathType.GetVariable().(int32)
			if gc.deathType < 0 || gc.deathType > 15 {
				gc.deathType = 0
			}
		}
	}
}

func (gc *GameCapture) Run() {
	for !ready {
		time.Sleep(time.Second)
	}
	for {
		if validVersion == false {
			return
		}
		gc.GetGameVariables()
		time.Sleep(time.Second / captureFPS)
	}
}
