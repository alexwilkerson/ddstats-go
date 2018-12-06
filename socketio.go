package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/wedeploy/gosocketio"
	"github.com/wedeploy/gosocketio/websocket"
)

const (
	sioStatusDisconnected = iota
	sioStatusConnecting
	sioStatusConnected
	sioStatusLoggedIn
	sioStatusTimeout
)

const sioTimeoutAttempts = 60

type SioVariables struct {
	status          int
	playerID        int
	timer           float32
	totalGems       int
	homing          int
	enemiesAlive    int
	enemiesKilled   int
	daggersHit      int
	daggersFired    int
	level2time      float32
	level3time      float32
	level4time      float32
	isReplay        bool
	deathType       int
	deathScreenSent bool
}

func (siov *SioVariables) Update() {
	siov.playerID = gameCapture.playerID
	siov.timer = gameCapture.timer
	siov.totalGems = gameCapture.totalGems
	siov.homing = gameCapture.homing
	siov.enemiesAlive = gameCapture.enemiesAlive
	siov.enemiesKilled = gameCapture.enemiesKilled
	siov.daggersHit = gameCapture.daggersHit
	siov.daggersFired = gameCapture.daggersFired
	siov.level2time = gameCapture.level2time
	siov.level3time = gameCapture.level3time
	siov.level4time = gameCapture.level4time
	siov.isReplay = gameCapture.isReplay
	if gameCapture.GetStatus() == statusIsDead {
		siov.deathType = gameCapture.deathType
	} else {
		siov.deathType = -1
	}
}

func liveStreamStats() {
	u := url.URL{
		Scheme: "ws",
		Host:   "ddstats.com",
	}

	var c *gosocketio.Client
	var err error
	for {

		if sioVariables.status == sioStatusDisconnected {
			for i := 0; i < sioTimeoutAttempts; i++ {
				debug.Log(fmt.Sprintf("Attempt %d connecting to server.", i+1))
				c, err = gosocketio.Connect(u, websocket.NewTransport())
				if err != nil {
					sioVariables.status = sioStatusConnecting
					debug.Log("Error connecting to server.")
					time.Sleep(time.Second)
					continue
				}
				break
			}
		}

		if sioVariables.status == sioStatusConnecting {
			debug.Log("Connection to server timed out.")
			sioVariables.status = sioStatusTimeout
			return
		}

		if err := c.On(gosocketio.OnDisconnect, sioDisconnectHandler); err != nil {
			return
		}

		if err := c.On(gosocketio.OnError, sioErrorHandler); err != nil {
			return
		}

		sioVariables.status = sioStatusConnected
		debug.Log("Connected to server.")

		// wait until a username has been stored before logging in.
		for gameCapture.GetStatus() == statusConnecting || gameCapture.GetStatus() == statusNotConnected || gameCapture.playerID > 1000000 {
			if sioVariables.status == sioStatusDisconnected {
				break
			}
			debug.Log("sio checking if dd.exe is connected.")
			time.Sleep(time.Second)
		}

		if sioVariables.status == sioStatusDisconnected {
			continue
		}

		// Allow time to fetch userID from server
		// there might be a safer way to do this.
		time.Sleep(time.Second)

		debug.Log(gameCapture.playerName)
		debug.Log(gameCapture.playerID)

		if sioVariables.status == sioStatusConnected {
			c.Emit("login", gameCapture.playerID)
		} else {
			continue
		}

		sioVariables.status = sioStatusLoggedIn

		for {
			if sioVariables.status == sioStatusDisconnected || gameCapture.GetStatus() == statusNotConnected || gameCapture.GetStatus() == statusConnecting {
				break
			}
			sioSubmit(*&c)
			time.Sleep(time.Second / sioFPS)
		}
	}
}

func sioSubmit(c *gosocketio.Client) {
	if gameCapture.GetStatus() == statusIsPlaying || (gameCapture.GetStatus() == statusIsDead && sioVariables.deathScreenSent == false) {
		if gameCapture.GetStatus() == statusIsDead {
			sioVariables.deathScreenSent = true
		} else if gameCapture.GetStatus() == statusIsPlaying {
			sioVariables.deathScreenSent = false
		}
		if err := c.Emit(
			"submit",
			sioVariables.playerID,
			sioVariables.timer,
			sioVariables.totalGems,
			sioVariables.homing,
			sioVariables.enemiesAlive,
			sioVariables.enemiesKilled,
			sioVariables.daggersHit,
			sioVariables.daggersFired,
			sioVariables.level2time,
			sioVariables.level3time,
			sioVariables.level4time,
			sioVariables.isReplay,
			sioVariables.deathType,
		); err != nil {
			return
		}
	}
}

func sioErrorHandler(err error) {
	debug.Log(err.Error())
	sioVariables.status = sioStatusDisconnected
}

func sioDisconnectHandler() {
	debug.Log("Disconnected.")
	sioVariables.status = sioStatusDisconnected
}
