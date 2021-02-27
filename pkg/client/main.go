package client

import (
	"fmt"
	"time"

	"github.com/alexwilkerson/ddstats-go/pkg/api"
	"github.com/alexwilkerson/ddstats-go/pkg/config"
	"github.com/alexwilkerson/ddstats-go/pkg/consoleui"
	"github.com/alexwilkerson/ddstats-go/pkg/devildaggers"
	"github.com/atotto/clipboard"
	ui "github.com/gizak/termui"
)

const (
	defaultTickRate   = time.Second / 36
	defaultUITickRate = time.Second / 2
)

type Client struct {
	version             string
	tickRate            time.Duration
	uiTickRate          time.Duration
	cfg                 *config.Config
	ui                  *consoleui.ConsoleUI
	uiData              *consoleui.Data
	dd                  *devildaggers.DevilDaggers
	apiClient           *api.Client
	loggedIn            bool
	lastSubmittedGameID int
	errChan             chan error
	done                chan struct{}
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
		Host:            cfg.Host,
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
	defer c.dd.StopPersistentConnection()

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
				c.copyGameURLToClipboard()
			}
		case err := <-c.errChan:
			close(c.done)
			return fmt.Errorf("Run: error returned on error channel: %w", err)
		}
	}
}

func (c *Client) run() {

	c.dd.StartPersistentConnection(c.errChan)
	go c.runDD()
	go c.runUI()
}

func (c *Client) runDD2() {
	for {
		select {
		case <-time.After(c.tickRate):
			if !c.dd.CheckConnection() {
				c.clearUIData()
				c.uiData.Status = consoleui.StatusDevilDaggersNotFound
				continue
			}
			c.populateUIData()
		case <-c.done:
			return
		}
	}
}

func (c *Client) runDD() {
	var oldStatus int32
	var statsSent bool
	for {
		select {
		case <-time.After(c.tickRate):
			if !c.dd.CheckConnection() {
				c.clearUIData()
				c.uiData.Status = consoleui.StatusDevilDaggersNotFound
				continue
			}

			c.populateUIData()

			newStatus := c.dd.GetStatus()

			// when a new game has started
			if oldStatus != devildaggers.StatusPlaying && newStatus == devildaggers.StatusPlaying ||
				oldStatus != devildaggers.StatusOtherReplay && newStatus == devildaggers.StatusOtherReplay ||
				oldStatus != devildaggers.StatusOwnReplayFromLeaderboard && newStatus == devildaggers.StatusOwnReplayFromLeaderboard {
				statsSent = false
			}

			if c.dd.GetStatsFinishedLoading() && !statsSent {
				if newStatus == devildaggers.StatusDead || newStatus == devildaggers.StatusOtherReplay || newStatus == devildaggers.StatusOwnReplayFromLeaderboard {
					// send stats
					submitGameInput, err := c.compileGameRecording()
					if err != nil {
						c.errChan <- fmt.Errorf("runGameCapture: could not compile game recording: %w", err)
					}
					gameID, err := c.apiClient.SubmitGame(submitGameInput)
					if err != nil {
						c.errChan <- fmt.Errorf("runGameCapture: error submitting game to server: %w", err)
					}
					c.lastSubmittedGameID = gameID
					statsSent = true
				}
			}

			oldStatus = newStatus
		case <-c.done:
			return
		}
	}
	// for {
	// 	select {
	// 	case <-time.After(c.tickRate):
	// 		if !c.dd.CheckConnection() {
	// 			oldStatus = devildaggers.StatusTitle
	// 			oldTime = 0.0
	// 			continue
	// 		}
	// 		newStatus := c.dd.GetStatus()
	// 		newTime := c.dd.GetTime()
	// 		if newStatus == devildaggers.StatusPlaying {
	// 			if newTime < 1 && (newStatus != oldStatus || oldTime > newTime) {
	// 				gameRecording, err = *c.newGameRecording()
	// 			}
	// 			if int(newTime)-int(gameRecording.TimerSlice[len(gameRecording.TimerSlice)-1]) >= 1 {
	// 				c.appendGameState(&gameRecording)
	// 			}
	// 			c.updateGameMaxValues(&gameRecording)
	// 		}
	// 		oldTime = newTime
	// 		oldStatus = newStatus
	// 	case <-c.done:
	// 		return
	// 	}
	// }
}

func (c *Client) compileGameRecording() (*api.SubmitGameInput, error) {
	submitGameInput := api.SubmitGameInput{
		PlayerID:            c.dd.GetPlayerID(),
		PlayerName:          c.dd.GetPlayerName(),
		Granularity:         1,
		Timer:               c.dd.GetTime(),
		TimerSlice:          []float32{},
		TotalGemsSlice:      []int32{},
		Level2time:          c.dd.GetTimeLvl2(),
		Level3time:          c.dd.GetTimeLvl3(),
		Level4time:          c.dd.GetTimeLvl4(),
		LeviDownTime:        c.dd.GetLeviathanDownTime(),
		OrbDownTime:         c.dd.GetOrbDownTime(),
		HomingSlice:         []int32{},
		HomingMax:           c.dd.GetHomingMax(),
		HomingMaxTime:       c.dd.GetHomingMaxTime(),
		DaggersFired:        c.dd.GetDaggersFired(),
		DaggersFiredSlice:   []int32{},
		DaggersHit:          c.dd.GetDaggersHit(),
		DaggersHitSlice:     []int32{},
		EnemiesAlive:        c.dd.GetEnemiesAlive(),
		EnemiesAliveSlice:   []int32{},
		EnemiesAliveMax:     c.dd.GetEnemiesAliveMax(),
		EnemiesAliveMaxTime: c.dd.GetEnemiesAliveMaxTime(),
		EnemiesKilled:       c.dd.GetKills(),
		EnemiesKilledSlice:  []int32{},
		DeathType:           c.dd.GetDeathType(),
		ReplayPlayerID:      c.dd.GetReplayPlayerID(),
		Version:             c.version,
		SurvivalHash:        c.dd.GetLevelHashMD5(),
	}

	statsFrame, err := c.dd.GetStatsFrame()
	if err != nil {
		return nil, fmt.Errorf("newGameRecording: could not refresh stats frame: %w", err)
	}

	for i, sf := range statsFrame {
		if i == len(statsFrame)-1 {
			submitGameInput.TimerSlice = append(submitGameInput.TimerSlice, c.dd.GetTime())
		} else {
			submitGameInput.TimerSlice = append(submitGameInput.TimerSlice, c.dd.GetStartingTime()+float32(i))
		}
		submitGameInput.TotalGemsSlice = append(submitGameInput.TotalGemsSlice, sf.TotalGems)
		submitGameInput.HomingSlice = append(submitGameInput.HomingSlice, sf.HomingDaggers)
		submitGameInput.DaggersFiredSlice = append(submitGameInput.DaggersFiredSlice, sf.DaggersFired)
		submitGameInput.DaggersHitSlice = append(submitGameInput.DaggersHitSlice, sf.DaggersHit)
		submitGameInput.EnemiesAliveSlice = append(submitGameInput.EnemiesAliveSlice, sf.EnemiesAlive)
		submitGameInput.EnemiesKilledSlice = append(submitGameInput.EnemiesKilledSlice, sf.Kills)
	}

	lastFrame := statsFrame[len(statsFrame)-1]

	submitGameInput.TotalGems = lastFrame.TotalGems
	submitGameInput.Homing = lastFrame.HomingDaggers

	return &submitGameInput, nil
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
			c.ui.ClearScreen()
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
	if c.uiData.PlayerName == "" {
		c.uiData.Status = consoleui.StatusConnecting
		return
	}
	c.uiData.LastGameID = c.lastSubmittedGameID
	status := c.dd.GetStatus()
	if status == devildaggers.StatusPlaying || status == devildaggers.StatusOtherReplay || status == devildaggers.StatusOwnReplayFromLastRun || status == devildaggers.StatusOwnReplayFromLeaderboard {
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

func (c *Client) copyGameURLToClipboard() {
	if c.lastSubmittedGameID != 0 {
		clipboard.WriteAll(fmt.Sprintf("%s/games/%d", c.cfg.Host, c.lastSubmittedGameID))
	}
}
