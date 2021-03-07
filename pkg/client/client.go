package client

import (
	"fmt"
	"time"

	"github.com/alexwilkerson/ddstats-go/pkg/config"
	"github.com/alexwilkerson/ddstats-go/pkg/consoleui"
	"github.com/alexwilkerson/ddstats-go/pkg/devildaggers"
	"github.com/alexwilkerson/ddstats-go/pkg/grpcclient"
	pb "github.com/alexwilkerson/ddstats-server/gamesubmission"
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
	grpcClient          *grpcclient.Client
	loggedIn            bool
	statsSent           bool
	lastSubmittedGameID int
	errChan             chan error
	done                chan struct{}
}

func New(version string, grpcAddr string) (*Client, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("New: unable to get config: %w", err)
	}

	grpcClient, err := grpcclient.New(grpcAddr)
	if err != nil {
		return nil, fmt.Errorf("New: unable to initialize grpc client: %w", err)
	}

	clientConnectReply, err := grpcClient.ClientConnect(version)
	if err != nil {
		return nil, fmt.Errorf("New: unable to connect to server: %w", err)
	}

	// TODO: handle invalid versions

	uiData := consoleui.Data{
		Host:            cfg.Host,
		MOTD:            clientConnectReply.GetMotd(),
		UpdateAvailable: clientConnectReply.UpdateAvailable,
		ValidVersion:    clientConnectReply.ValidVersion,
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
		grpcClient: grpcClient,
		statsSent:  true, // this is true to prevent stats being sent when game is opened while on death screen
		errChan:    make(chan error),
		done:       make(chan struct{}),
	}, nil
}

// Run starts the client.
func (c *Client) Run() error {
	defer c.ui.Close()
	defer c.dd.StopPersistentConnection()
	defer c.grpcClient.Close()

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
				c.statsSent = false
			}

			if c.dd.GetStatsFinishedLoading() && !c.statsSent {
				if newStatus == devildaggers.StatusDead || newStatus == devildaggers.StatusOtherReplay || newStatus == devildaggers.StatusOwnReplayFromLeaderboard {
					// send stats
					submitGameRequest, err := c.compileGameRequest()
					if err != nil {
						c.errChan <- fmt.Errorf("runGameCapture: could not compile game recording: %w", err)
					}
					gameID, err := c.grpcClient.SubmitGame(submitGameRequest)
					if err != nil {
						c.errChan <- fmt.Errorf("runGameCapture: error submitting game to server: %w", err)
					}
					c.lastSubmittedGameID = gameID
					c.statsSent = true
				}
			}

			oldStatus = newStatus
		case <-c.done:
			return
		}
	}
}

func (c *Client) compileGameRequest() (*pb.SubmitGameRequest, error) {
	playerID := c.dd.GetPlayerID()
	var replayPlayerID int32
	if c.dd.GetIsReplay() {
		playerID = c.dd.GetReplayPlayerID()
		replayPlayerID = c.dd.GetPlayerID()
	}
	submitGameRequest := pb.SubmitGameRequest{
		Version:              c.version,
		PlayerID:             playerID,
		PlayerName:           c.dd.GetPlayerName(),
		LevelHashMD5:         c.dd.GetLevelHashMD5(),
		TimeLvll2:            c.dd.GetTimeLvl2(),
		TimeLvll3:            c.dd.GetTimeLvl3(),
		TimeLvll4:            c.dd.GetTimeLvl4(),
		TimeLeviDown:         c.dd.GetLeviathanDownTime(),
		TimeOrbDown:          c.dd.GetOrbDownTime(),
		EnemiesAliveMax:      c.dd.GetEnemiesAliveMax(),
		EnemiesAliveMaxTime:  c.dd.GetEnemiesAliveMaxTime(),
		HomingDaggersMax:     c.dd.GetHomingMax(),
		HomingDaggersMaxTime: c.dd.GetHomingMaxTime(),
		DeathType:            uint32(c.dd.GetDeathType()),
		IsReplay:             c.dd.GetIsReplay(),
		ReplayPlayerID:       replayPlayerID,
		Stats:                []*pb.StatFrame{},
	}

	statsFrame, err := c.dd.GetStatsFrame()
	if err != nil {
		return nil, fmt.Errorf("newGameRecording: could not refresh stats frame: %w", err)
	}

	for _, sf := range statsFrame {
		perEnemyAliveCount := make([]int32, len(sf.PerEnemyAliveCount))
		for i := range sf.PerEnemyAliveCount {
			perEnemyAliveCount[i] = int32(sf.PerEnemyAliveCount[i])
		}
		perEnemyKillCount := make([]int32, len(sf.PerEnemyKillCount))
		for i := range sf.PerEnemyKillCount {
			perEnemyKillCount[i] = int32(sf.PerEnemyKillCount[i])
		}
		submitGameRequest.Stats = append(submitGameRequest.Stats, &pb.StatFrame{
			GemsCollected:      sf.GemsCollected,
			Kills:              sf.Kills,
			DaggersFired:       sf.DaggersFired,
			DaggersHit:         sf.DaggersHit,
			EnemiesAlive:       sf.EnemiesAlive,
			LevelGems:          sf.LevelGems,
			HomingDaggers:      sf.HomingDaggers,
			GemsDespawned:      sf.GemsDespawned,
			GemsEaten:          sf.GemsEaten,
			TotalGems:          sf.TotalGems,
			DaggersEaten:       sf.DaggersEaten,
			PerEnemyAliveCount: perEnemyAliveCount,
			PerEnemyKillCount:  perEnemyKillCount,
		})
	}

	lastFrame := submitGameRequest.Stats[len(statsFrame)-1]

	submitGameRequest.GemsCollected = lastFrame.GemsCollected
	submitGameRequest.Kills = lastFrame.Kills
	submitGameRequest.DaggersFired = lastFrame.DaggersFired
	submitGameRequest.DaggersHit = lastFrame.DaggersHit
	submitGameRequest.EnemiesAlive = lastFrame.EnemiesAlive
	submitGameRequest.LevelGems = lastFrame.LevelGems
	submitGameRequest.HomingDaggers = lastFrame.HomingDaggers
	submitGameRequest.GemsDespawned = lastFrame.GemsCollected
	submitGameRequest.GemsEaten = lastFrame.GemsEaten
	submitGameRequest.TotalGems = lastFrame.TotalGems
	submitGameRequest.DaggersEaten = lastFrame.DaggersEaten
	submitGameRequest.PerEnemyAliveCount = lastFrame.PerEnemyAliveCount
	submitGameRequest.PerEnemyKillcount = lastFrame.PerEnemyKillCount
	submitGameRequest.Time = c.dd.GetTimeMax()

	return &submitGameRequest, nil
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
	c.uiData.Recording = consoleui.StatusNotRecording
	c.uiData.Timer = 0.0
	c.uiData.DaggersHit = 0
	c.uiData.DaggersFired = 0
	c.uiData.Accuracy = 0.0
	c.uiData.TotalGems = 0
	c.uiData.Homing = 0
	c.uiData.EnemiesAlive = 0
	c.uiData.EnemiesKilled = 0
	c.uiData.GemsDespawned = 0
	c.uiData.GemsEaten = 0
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
		c.uiData.Recording = consoleui.StatusRecording
		if c.statsSent {
			c.uiData.Recording = consoleui.StatusGameSubmitted
		}
		c.uiData.Timer = c.dd.GetTime()
		c.uiData.DaggersHit = c.dd.GetDaggersHit()
		c.uiData.DaggersFired = c.dd.GetDaggersFired()
		c.uiData.Accuracy = c.dd.GetAccuracy()
		c.uiData.TotalGems = c.dd.GetGemsCollected()
		c.uiData.Homing = c.dd.GetHomingDaggers()
		c.uiData.EnemiesAlive = c.dd.GetEnemiesAlive()
		c.uiData.EnemiesKilled = c.dd.GetKills()
		c.uiData.GemsDespawned = c.dd.GetGemsDespawned()
		c.uiData.GemsEaten = c.dd.GetGemsEaten()
	} else {
		c.uiData.Recording = consoleui.StatusNotRecording
		if c.dd.GetStatus() == devildaggers.StatusDead {
			if c.statsSent {
				c.uiData.Recording = consoleui.StatusGameSubmitted
			}
			c.uiData.DeathType = c.dd.GetDeathType()
		}
	}
}

func (c *Client) copyGameURLToClipboard() {
	if c.lastSubmittedGameID != 0 {
		clipboard.WriteAll(fmt.Sprintf("%s/games/%d", c.cfg.Host, c.lastSubmittedGameID))
	}
}
