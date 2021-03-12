package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/alexwilkerson/ddstats-go/pkg/config"
	"github.com/alexwilkerson/ddstats-go/pkg/consoleui"
	"github.com/alexwilkerson/ddstats-go/pkg/devildaggers"
	"github.com/alexwilkerson/ddstats-go/pkg/grpcclient"
	"github.com/alexwilkerson/ddstats-go/pkg/socketio"
	pb "github.com/alexwilkerson/ddstats-server/gamesubmission"
	"github.com/atotto/clipboard"
	ui "github.com/gizak/termui"
)

const (
	defaultTickRate    = time.Second / 36
	defaultUITickRate  = time.Second / 2
	defaultSIOTickRate = time.Second / 3
)

type Client struct {
	version             string
	v3SurvivalHash      string
	tickRate            time.Duration
	uiTickRate          time.Duration
	cfg                 *config.Config
	ui                  *consoleui.ConsoleUI
	uiData              *consoleui.Data
	dd                  *devildaggers.DevilDaggers
	grpcClient          *grpcclient.Client
	sioClient           *socketio.Client
	loggedIn            bool
	statsSent           bool
	lastSubmittedGameID int
	errChan             chan error
	done                chan struct{}
}

func New(version string, grpcAddr, v3SurvivalHash string) (*Client, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("New: unable to get config: %w", err)
	}

	grpcClient, err := grpcclient.New(grpcAddr)
	if err != nil {
		return nil, fmt.Errorf("New: unable to initialize grpc client: %w", err)
	}

	motd, updateAvailable, validVersion := "Offline Mode", false, false

	if !cfg.OfflineMode && (cfg.GetMOTD || cfg.CheckForUpdates) {
		clientConnectReply, err := grpcClient.ClientConnect(version)
		if err != nil {
			return nil, fmt.Errorf("New: unable to connect to server: %w", err)
		}
		if cfg.GetMOTD {
			motd = clientConnectReply.GetMotd()
		} else {
			motd = ""
		}
		if cfg.CheckForUpdates {
			updateAvailable = clientConnectReply.UpdateAvailable
		}
		validVersion = clientConnectReply.ValidVersion

		if !validVersion {
			return nil, errors.New("invalid version: tell mother")
		}
	}

	sioClient, err := socketio.New(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("New: unable to connect to socketio: %w", err)
	}

	// TODO: handle invalid versions

	uiData := consoleui.Data{
		Host:            cfg.Host,
		MOTD:            motd,
		UpdateAvailable: updateAvailable,
		ValidVersion:    validVersion,
		Version:         version,
	}

	ui, err := consoleui.New(&uiData)
	if err != nil {
		return nil, fmt.Errorf("New: could not create ui: %w", err)
	}

	dd := devildaggers.New()

	return &Client{
		version:        version,
		v3SurvivalHash: v3SurvivalHash,
		tickRate:       defaultTickRate,
		uiTickRate:     defaultUITickRate,
		cfg:            cfg,
		ui:             ui,
		uiData:         &uiData,
		dd:             dd,
		grpcClient:     grpcClient,
		sioClient:      sioClient,
		statsSent:      true, // this is true to prevent stats being sent when game is opened while on death screen
		errChan:        make(chan error),
		done:           make(chan struct{}),
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
	if !c.cfg.OfflineMode {
		go c.runSIO()
	}
}

func (c *Client) runSIO() {
	defer func() {
		if c.sioClient.GetStatus() != socketio.StatusDisconnected {
			err := c.sioClient.Disconnect()
			if err != nil {
				c.errChan <- fmt.Errorf("runSIO: error disconnecting from sio: %w", err)
				return
			}
		}
	}()
	for {
		select {
		case <-time.After(defaultSIOTickRate):
			if c.dd.CheckConnection() {
				if c.sioClient.GetStatus() != socketio.StatusLoggedIn {
					if c.dd.GetPlayerID() != 0 {
						err := c.sioClient.Connect(int(c.dd.GetPlayerID()))
						if err != nil {
							c.errChan <- fmt.Errorf("runSIO: error connecting to sio: %w", err)
							return
						}
					}
				} else {
					if c.dd.GetIsInGame() || c.dd.GetStatus() == devildaggers.StatusDead {
						if (c.cfg.Stream.Stats && !c.dd.GetIsReplay()) ||
							(c.cfg.Stream.ReplayStats && c.dd.GetIsReplay()) {
							if (c.dd.GetLevelHashMD5() == c.v3SurvivalHash) ||
								(!c.cfg.Stream.NonDefaultSpawnsets && c.dd.GetLevelHashMD5() != c.v3SurvivalHash) {
								var deathType int32 = -2
								if c.dd.GetStatus() == devildaggers.StatusPlaying {
									deathType = -1
								} else if c.dd.GetStatus() == devildaggers.StatusDead {
									deathType = int32(c.dd.GetDeathType())
								}

								err := c.sioClient.SubmitStats(&socketio.SubmissionData{
									PlayerID:         c.dd.GetPlayerID(),
									Timer:            c.dd.GetTime(),
									TotalGems:        c.dd.GetGemsCollected(),
									Homing:           c.dd.GetHomingDaggers(),
									EnemiesAlive:     c.dd.GetEnemiesAlive(),
									EnemiesKilled:    c.dd.GetKills(),
									DaggersHit:       c.dd.GetDaggersHit(),
									DaggersFired:     c.dd.GetDaggersFired(),
									Level2time:       c.dd.GetTimeLvl2(),
									Level3time:       c.dd.GetTimeLvl3(),
									Level4time:       c.dd.GetTimeLvl4(),
									IsReplay:         c.dd.GetIsReplay(),
									DeathType:        deathType,
									NotifyPlayerBest: c.cfg.Discord.NotifyPlayerBest,
									NotifyAbove1000:  c.cfg.Discord.NotifyAbove1100,
								})
								if err != nil {
									c.errChan <- fmt.Errorf("runSIO: error sending stats via sio: %w", err)
									return
								}
							}
						}
					} else {
						var sioStatus int
						switch c.dd.GetStatus() {
						case devildaggers.StatusTitle, devildaggers.StatusMenu:
							sioStatus = 4
						case devildaggers.StatusLobby:
							sioStatus = 5
						case devildaggers.StatusPlaying:
							sioStatus = 2
						case devildaggers.StatusDead:
							sioStatus = 6
						case devildaggers.StatusOwnReplayFromLastRun, devildaggers.StatusOwnReplayFromLeaderboard, devildaggers.StatusOtherReplay:
							sioStatus = 3
						}

						err := c.sioClient.SubmitStatusUpdate(int(c.dd.GetPlayerID()), sioStatus)
						if err != nil {
							c.errChan <- fmt.Errorf("runSIO: error sending status update via sio: %w", err)
							return
						}
					}
				}
			} else {
				if c.sioClient.GetStatus() == socketio.StatusLoggedIn {
					err := c.sioClient.Disconnect()
					if err != nil {
						c.errChan <- fmt.Errorf("runSIO: error disconnecting from sio: %w", err)
						return
					}
				}
			}
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
				c.uiData.OnlineStatus = c.sioClient.GetStatus()
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

			if !c.cfg.OfflineMode {
				if c.dd.GetStatsFinishedLoading() && !c.statsSent {
					if newStatus == devildaggers.StatusDead || newStatus == devildaggers.StatusOtherReplay || newStatus == devildaggers.StatusOwnReplayFromLeaderboard {
						// send stats
						submitGameRequest, err := c.compileGameRequest()
						if err != nil {
							c.errChan <- fmt.Errorf("runGameCapture: could not compile game recording: %w", err)
							return
						}
						gameID, err := c.grpcClient.SubmitGame(submitGameRequest)
						if err != nil {
							c.errChan <- fmt.Errorf("runGameCapture: error submitting game to server: %w", err)
							return
						}
						c.lastSubmittedGameID = gameID
						c.statsSent = true

						if c.cfg.AutoClipboardGame {
							c.copyGameURLToClipboard()
						}

						if (c.cfg.Submit.Stats && !c.dd.GetIsReplay()) ||
							(c.cfg.Submit.ReplayStats && c.dd.GetIsReplay()) {
							if (c.dd.GetLevelHashMD5() == c.v3SurvivalHash) ||
								(!c.cfg.Submit.NonDefaultSpawnsets && c.dd.GetLevelHashMD5() != c.v3SurvivalHash) {
								if c.sioClient.GetStatus() == socketio.StatusLoggedIn {
									err = c.sioClient.SubmitGame(gameID, c.cfg.Discord.NotifyPlayerBest, c.cfg.Discord.NotifyAbove1100)
									if err != nil {
										c.errChan <- fmt.Errorf("runGameCapture: error submitting game to sio: %w", err)
										return
									}
								}
							}
						}
					}
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
		TimeLvl2:             c.dd.GetTimeLvl2(),
		TimeLvl3:             c.dd.GetTimeLvl3(),
		TimeLvl4:             c.dd.GetTimeLvl4(),
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
	c.uiData.OnlineStatus = c.sioClient.GetStatus()
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
		c.uiData.GemsCollected = c.dd.GetGemsCollected()
		c.uiData.Homing = c.dd.GetHomingDaggers()
		c.uiData.EnemiesAlive = c.dd.GetEnemiesAlive()
		c.uiData.EnemiesKilled = c.dd.GetKills()
		c.uiData.TotalGems = c.dd.GetTotalGems()
		c.uiData.GemsDespawned = c.dd.GetGemsDespawned()
		c.uiData.GemsEaten = c.dd.GetGemsEaten()
		c.uiData.DaggersEaten = c.dd.GetDaggersEaten()
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
