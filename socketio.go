package main

import (
	"net/url"
	"time"

	"github.com/wedeploy/gosocketio"
	"github.com/wedeploy/gosocketio/websocket"
)

type SioVariables struct {
	online          bool
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

	c, err := gosocketio.Connect(u, websocket.NewTransport())
	if err != nil {
		return
	}

	// wait until a username has been stored before logging in.
	for gameCapture.playerName == "" {
		time.Sleep(time.Second)
	}

	c.Emit("login", gameCapture.playerID)

	if err := c.On(gosocketio.OnError, sioErrorHandler); err != nil {
		return
	}

	sioVariables.online = true

	if err := c.On(gosocketio.OnDisconnect, sioDisconnectHandler); err != nil {
		return
	}

	for {
		sioSubmit(*&c)
		time.Sleep(time.Second / sioFPS)
	}

}

func sioSubmit(c *gosocketio.Client) {
	if gameCapture.GetStatus() == statusIsDead {
		if sioVariables.deathScreenSent == false {
			sioVariables.deathScreenSent = true
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
			return
		} else {
			return
		}
	}
	sioVariables.deathScreenSent = false
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

func sioErrorHandler(err error) {
	debug.Log(err.Error())
	sioVariables.online = false
}

func sioDisconnectHandler() {
	debug.Log("Disconnected.")
	sioVariables.online = false
}
