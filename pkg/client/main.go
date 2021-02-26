package client

import (
	"fmt"
	"time"

	"github.com/alexwilkerson/ddstats-go/pkg/api"
	"github.com/alexwilkerson/ddstats-go/pkg/config"
	"github.com/alexwilkerson/ddstats-go/pkg/consoleui"
	"github.com/alexwilkerson/ddstats-go/pkg/devildaggers"
	ui "github.com/gizak/termui"
)

const (
	defaultTickRate   = time.Second / 36
	defaultUITickRate = time.Second / 2
)

type Client struct {
	version    string
	tickRate   time.Duration
	uiTickRate time.Duration
	cfg        *config.Config
	ui         *consoleui.ConsoleUI
	uiData     *consoleui.Data
	dd         *devildaggers.DevilDaggers
	apiClient  *api.Client
	loggedIn   bool
	errChan    chan error
	done       chan struct{}
}

func New(version string) (*Client, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("New: unable to get config: %w", err)
	}

	apiClient, err := api.New(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("New: could not create api client: %w", err)
	}

	resp, err := apiClient.InitConnection(version)
	if err != nil {
		return nil, fmt.Errorf("New: unable to initialize connection: %w", err)
	}

	// TODO: handle invalid versions

	uiData := consoleui.Data{
		MOTD:            resp.MOTD,
		UpdateAvailable: resp.UpdateAvailable,
		Version:         version,
	}

	ui, err := consoleui.New(&uiData)
	if err != nil {
		return nil, fmt.Errorf("New: could not create ui: %w", err)
	}

	dd := devildaggers.New()

	return &Client{
		version:    version,
		tickRate:   defaultTickRate,
		uiTickRate: defaultUITickRate,
		cfg:        cfg,
		ui:         ui,
		uiData:     &uiData,
		dd:         dd,
		apiClient:  apiClient,
		errChan:    make(chan error),
		done:       make(chan struct{}),
	}, nil
}

func (c *Client) Run() error {
	defer c.ui.Close()
	defer c.dd.Close()

	go c.run()

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>", "<f10>":
				close(c.done)
				return nil
			case "<f12>":
				config.WriteDefaultConfigFile()
			case "<MouseLeft>":
				copyGameURLToClipboard()
			}
		case err := <-c.errChan:
			return fmt.Errorf("Run: error returned on error channel: %w", err)
		}
	}
}

func (c *Client) run() {

	c.dd.StartPersistentConnection(c.errChan)
	go c.runDD()
	go c.runUI()
}

func (c *Client) runDD() {
	for {
		select {
		case <-time.After(c.tickRate):
			if !c.dd.CheckConnection() {
				_, err := c.dd.Connect()
				if err != nil {
					c.uiData.Status = consoleui.StatusDevilDaggersNotFound
					c.clearUIData()
					continue
				}
			}
			// if !c.loggedIn {

			// }
			err := c.dd.RefreshData()
			if err != nil {
				c.uiData.Status = consoleui.StatusDevilDaggersNotFound
				continue
			}
			c.populateUIData()
		case <-c.done:
			return
		}
	}
}

func (c *Client) runGameCapture() {
	var gameRecording api.SubmitGameInput
	oldStatus := devildaggers.StatusTitle
	var oldTime float32
	for {
		select {
		case <-time.After(c.tickRate):
			if !c.dd.CheckConnection() {
				oldStatus = devildaggers.StatusTitle
				oldTime = 0.0
				continue
			}
			newStatus := c.dd.GetStatus()
			newTime := c.dd.GetTime()
			if newStatus == devildaggers.StatusPlaying {
				if newTime < 1 && (newStatus != oldStatus || oldTime > newTime) {
					gameRecording = *c.newGameRecording()
				}
				if int(newTime)-int(gameRecording.TimerSlice[len(gameRecording.TimerSlice)-1]) >= 1 {
					c.appendGameState(&gameRecording)
				}
				c.updateGameMaxValues(&gameRecording)
			}
			oldTime = newTime
			oldStatus = newStatus
		case <-c.done:
			return
		}
	}
}

func (c *Client) newGameRecording() *api.SubmitGameInput {
	return &api.SubmitGameInput{
		PlayerID:           c.dd.GetPlayerID(),
		PlayerName:         c.dd.GetPlayerName(),
		Granularity:        1,
		TimerSlice:         []float32{c.dd.GetTime()},
		TotalGemsSlice:     []uint32{c.dd.GetTotalGems()},
		HomingSlice:        []uint32{c.dd.GetHomingDaggers()},
		DaggersFiredSlice:  []uint32{c.dd.GetDaggersFired()},
		DaggersHitSlice:    []uint32{c.dd.GetDaggersHit()},
		EnemiesAliveSlice:  []uint32{c.dd.GetEnemiesAlive()},
		EnemiesKilledSlice: []uint32{c.dd.GetKills()},
		ReplayPlayerID:     c.dd.GetReplayPlayerID(),
		Version:            c.version,
		SurvivalHash:       c.dd.GetLevelHashMD5(),
	}
}

func (c *Client) updateGameMaxValues(gameRecording *api.SubmitGameInput) {
	time := c.dd.GetTime()
	totalGems := c.dd.GetTotalGems()
	homing := c.dd.GetHomingDaggers()
	daggersFired := c.dd.GetDaggersFired()
	daggersHit := c.dd.GetDaggersHit()
	enemiesAlive := c.dd.GetEnemiesAlive()
	enemiesKilled := c.dd.GetKills()

	if totalGems > gameRecording.TotalGemsSlice[len(gameRecording.TotalGemsSlice)-1] {
		gameRecording.TotalGemsSlice[len(gameRecording.TotalGemsSlice)-1] = totalGems
	}

	if homing > gameRecording.HomingMax {
		gameRecording.HomingMax = homing
		gameRecording.HomingMaxTime = time
	}
	if homing > gameRecording.HomingSlice[len(gameRecording.HomingSlice)-1] {
		gameRecording.HomingSlice[len(gameRecording.HomingSlice)-1] = homing
	}

	if daggersFired > gameRecording.DaggersFiredSlice[len(gameRecording.DaggersFiredSlice)-1] {
		gameRecording.DaggersFiredSlice[len(gameRecording.DaggersFiredSlice)-1] = daggersFired
	}

	if daggersHit > gameRecording.DaggersHitSlice[len(gameRecording.DaggersHitSlice)-1] {
		gameRecording.DaggersHitSlice[len(gameRecording.DaggersHitSlice)-1] = daggersHit
	}

	if enemiesAlive > gameRecording.EnemiesAliveMax {
		gameRecording.EnemiesAliveMax = enemiesAlive
		gameRecording.EnemiesAliveMaxTime = time
	}
	if enemiesAlive > gameRecording.EnemiesAliveSlice[len(gameRecording.EnemiesAliveSlice)-1] {
		gameRecording.EnemiesAliveSlice[len(gameRecording.EnemiesAliveSlice)-1] = enemiesAlive
	}

	if enemiesKilled > gameRecording.EnemiesKilledSlice[len(gameRecording.EnemiesKilledSlice)-1] {
		gameRecording.EnemiesKilledSlice[len(gameRecording.EnemiesKilledSlice)-1] = enemiesKilled
	}
}

func (c *Client) appendGameState(gameRecording *api.SubmitGameInput) {
	gameRecording.TimerSlice = append(gameRecording.TimerSlice, c.dd.GetTime())
	gameRecording.TotalGemsSlice = append(gameRecording.TotalGemsSlice, c.dd.GetTotalGems())
	gameRecording.HomingSlice = append(gameRecording.HomingSlice, c.dd.GetHomingDaggers())
	gameRecording.DaggersFiredSlice = append(gameRecording.DaggersFiredSlice, c.dd.GetDaggersFired())
	gameRecording.DaggersHitSlice = append(gameRecording.DaggersHitSlice, c.dd.GetDaggersHit())
	gameRecording.EnemiesAliveSlice = append(gameRecording.EnemiesAliveSlice, c.dd.GetEnemiesAlive())
	gameRecording.EnemiesKilledSlice = append(gameRecording.EnemiesKilledSlice, c.dd.GetKills())
}

func (c *Client) runUI() {
	c.ui.ClearScreen()
	for {
		select {
		case <-time.After(c.tickRate):
			err := c.ui.DrawScreen()
			if err != nil {
				c.errChan <- fmt.Errorf("runUI: error drawing screen in ui: %w", err)
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) clearUIData() {
	c.uiData.PlayerName = ""
	c.uiData.Recording = false
	c.uiData.Timer = 0.0
	c.uiData.DaggersHit = 0
	c.uiData.DaggersFired = 0
	c.uiData.Accuracy = 0.0
	c.uiData.TotalGems = 0
	c.uiData.Homing = 0
	c.uiData.EnemiesAlive = 0
	c.uiData.EnemiesKilled = 0
	c.uiData.DeathType = 0
}

func (c *Client) populateUIData() {
	c.uiData.Status = c.dd.GetStatus()
	c.uiData.PlayerName = c.dd.GetPlayerName()
	if c.dd.GetStatus() == devildaggers.StatusPlaying {
		c.uiData.Recording = true
		c.uiData.Timer = c.dd.GetTime()
		c.uiData.DaggersHit = c.dd.GetDaggersHit()
		c.uiData.DaggersFired = c.dd.GetDaggersFired()
		c.uiData.Accuracy = c.dd.GetAccuracy()
		c.uiData.TotalGems = c.dd.GetTotalGems()
		c.uiData.Homing = c.dd.GetHomingDaggers()
		c.uiData.EnemiesAlive = c.dd.GetEnemiesAlive()
		c.uiData.EnemiesKilled = c.dd.GetKills()
	} else {
		if c.dd.GetStatus() == devildaggers.StatusDead {
			c.uiData.DeathType = c.dd.GetDeathType()
		}
		c.uiData.Recording = false
	}
}

func copyGameURLToClipboard() {
	// if lastGameURL[:4] == "https" {
	// 	lastGameURLCopyTime = time.Now()
	// 	clipboard.WriteAll(lastGameURL)
	// }
}
