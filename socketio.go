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
	status          int32
	playerID        int32
	timer           float32
	totalGems       int32
	homing          int32
	enemiesAlive    int32
	enemiesKilled   int32
	daggersHit      int32
	daggersFired    int32
	level2time      float32
	level3time      float32
	level4time      float32
	isReplay        bool
	deathType       int32
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
	} else if gameCapture.GetStatus() == statusIsPlaying || gameCapture.GetStatus() == statusIsReplay {
		siov.deathType = -1
	} else {
		siov.deathType = -2
	}
}

var sioClient *gosocketio.Client

func liveStreamStats() {
	u := url.URL{
		Scheme: "ws",
		Host:   "ddstats.com",
	}

	var err error

	for {

		for gameCapture.GetStatus() == statusNotConnected {
			time.Sleep(time.Second * 2)
		}

		if sioVariables.status == sioStatusDisconnected {
			for i := 0; i < sioTimeoutAttempts; i++ {
				debug.Log(fmt.Sprintf("Attempt %d connecting to server.", i+1))
				sioClient, err = gosocketio.Connect(u, websocket.NewTransport())
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

		if err := sioClient.On(gosocketio.OnDisconnect, sioDisconnectHandler); err != nil {
			return
		}

		if err := sioClient.On(gosocketio.OnError, sioErrorHandler); err != nil {
			return
		}

		sioVariables.status = sioStatusConnected
		debug.Log("Connected to server.")

		// wait until a username has been stored before logging in.
		for gameCapture.GetStatus() == statusConnecting || gameCapture.GetStatus() == statusNotConnected || gameCapture.playerID == -1 {
			if sioVariables.status == sioStatusDisconnected {
				break
			}
			debug.Log(gameCapture.playerID)
			// debug.Log("sio checking if dd.exe is connected.")
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
			sioClient.Emit("login", gameCapture.playerID)
		} else {
			continue
		}

		sioVariables.status = sioStatusLoggedIn

		for {
			if sioVariables.status == sioStatusDisconnected || gameCapture.GetStatus() == statusNotConnected || gameCapture.GetStatus() == statusConnecting {
				sioClient.Close()
				break
			}
			err := sioSubmit(*&sioClient)
			if err != nil {
				sioClient.Close()
				break
			}
			time.Sleep(time.Second / sioFPS)
		}
	}
}

func sioSubmit(c *gosocketio.Client) error {
	if gameCapture.GetStatus() == statusIsPlaying ||
		gameCapture.GetStatus() == statusIsReplay ||
		(gameCapture.GetStatus() == statusIsDead && sioVariables.deathScreenSent == false) ||
		gameCapture.GetStatus() == statusInMainMenu {

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
			return err
		}
	}
	return nil
}

func sioErrorHandler(err error) {
	debug.Log(err.Error())
	sioVariables.status = sioStatusDisconnected
}

func sioDisconnectHandler() {
	debug.Log("Disconnected.")
	sioVariables.status = sioStatusDisconnected
}
