package socketio

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/wedeploy/gosocketio"
	"github.com/wedeploy/gosocketio/websocket"
)

const defaultTickRate = time.Second / 3

const (
	submitFuncName        = "submit"
	gameSubmittedFuncName = "game_submitted"
	statusUpdateFuncName  = "status_update"
	loginFuncName         = "login"
)

const (
	StatusDisconnected = iota
	StatusConnecting
	StatusConnected
	StatusLoggedIn
	StatusTimeout
)

type Client struct {
	status    int
	hostURL   *url.URL
	sioClient *gosocketio.Client
}

func New(hostURL string) (*Client, error) {
	parsedHostURL, err := url.Parse(hostURL)
	if err != nil {
		return nil, fmt.Errorf("New: could not parse host url: %w", err)
	}

	scheme := "wss"
	if parsedHostURL.Scheme != "https" {
		scheme = "ws"
	}

	u := url.URL{
		Scheme: scheme,
		Host:   parsedHostURL.Host,
	}

	return &Client{
		hostURL: &u,
	}, nil
}

func (c *Client) Connect(playerID int) error {
	var err error

	c.sioClient = &gosocketio.Client{}

	c.status = StatusConnecting

	c.sioClient, err = gosocketio.Connect(*c.hostURL, websocket.NewTransport())
	if err != nil {
		return fmt.Errorf("Connect: could not connec to host: %w", err)
	}

	c.status = StatusConnected

	err = c.sioClient.On(gosocketio.OnDisconnect, c.disconnectHandler)
	if err != nil {
		return fmt.Errorf("Connect: could not register disconnect handler")
	}

	err = c.sioClient.On(gosocketio.OnError, c.errorHandler)
	if err != nil {
		return fmt.Errorf("Connect: could not register error handler")
	}

	err = c.SubmitLogin(playerID)
	if err != nil {
		return fmt.Errorf("Connect: could not log in to server: %w", err)
	}

	c.status = StatusLoggedIn

	return nil
}

func (c *Client) Disconnect() error {
	if c.sioClient == nil {
		return nil
	}
	err := c.sioClient.Close()
	if err != nil {
		return fmt.Errorf("Disconnect: error disconnecting: %w", err)
	}
	c.status = StatusDisconnected
	return nil
}

func (c *Client) GetStatus() int {
	return c.status
}

type SubmissionData struct {
	PlayerID         int32
	Timer            float32
	TotalGems        int32
	Homing           int32
	EnemiesAlive     int32
	EnemiesKilled    int32
	DaggersHit       int32
	DaggersFired     int32
	Level2time       float32
	Level3time       float32
	Level4time       float32
	IsReplay         bool
	DeathType        int32
	NotifyPlayerBest bool
	NotifyAbove1000  bool
	DeathScreenSent  bool
}

func (c *Client) SubmitStats(submissionData *SubmissionData) error {
	if c.sioClient == nil {
		return errors.New("SubmitStats: sioClient is nil")
	}
	err := c.sioClient.Emit(
		submitFuncName,
		submissionData.PlayerID,
		submissionData.Timer,
		submissionData.TotalGems,
		submissionData.Homing,
		submissionData.EnemiesAlive,
		submissionData.EnemiesKilled,
		submissionData.DaggersHit,
		submissionData.DaggersFired,
		submissionData.Level2time,
		submissionData.Level3time,
		submissionData.Level4time,
		submissionData.IsReplay,
		submissionData.DeathType,
		submissionData.NotifyPlayerBest,
		submissionData.NotifyAbove1000,
	)
	if err != nil {
		return fmt.Errorf("SubmitStats: error submitting: %w", err)
	}
	return nil
}

func (c *Client) SubmitGame(gameID int, notifyPlayerBest, notifyAbove1000 bool) error {
	if c.sioClient == nil {
		return errors.New("SubmitGame: sioClient is nil")
	}
	err := c.sioClient.Emit(
		gameSubmittedFuncName,
		gameID,
		notifyPlayerBest,
		notifyAbove1000,
	)
	if err != nil {
		return fmt.Errorf("SubmitGame: error sending 'game_submitted' func via sio: %w", err)
	}
	return nil
}

func (c *Client) SubmitStatusUpdate(playerID int, status int) error {
	if c.sioClient == nil {
		return errors.New("SubmitStatusUpdate: sioClient is nil")
	}
	err := c.sioClient.Emit(
		statusUpdateFuncName,
		playerID,
		status,
	)
	if err != nil {
		return fmt.Errorf("SubmitStatusUpdate: error submitting status update: %w", err)
	}
	return nil
}

func (c *Client) SubmitLogin(playerID int) error {
	if c.sioClient == nil {
		return errors.New("SubmitLogin: sioClient is nil")
	}
	err := c.sioClient.Emit(
		loginFuncName,
		playerID,
	)
	if err != nil {
		return fmt.Errorf("SubmitLogin: error submitting login: %w", err)
	}
	return nil
}

func (c *Client) errorHandler(inputErr error) {
	// f, err := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	return
	// }
	// defer f.Close()

	// log.SetOutput(f)
	// log.Printf("%v\n", inputErr)
	c.status = StatusDisconnected
}

func (c *Client) disconnectHandler() {
	c.status = StatusDisconnected
}
