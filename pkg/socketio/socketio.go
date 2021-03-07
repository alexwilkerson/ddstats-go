package socketio

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/wedeploy/gosocketio"
	"github.com/wedeploy/gosocketio/websocket"
)

const (
	submitFuncName       = "submit"
	statusUpdateFuncName = "status_update"
	loginFuncName        = "login"
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
		hostURL:   &u,
		sioClient: &gosocketio.Client{},
	}, nil
}

func (c *Client) Connect(playerID int) error {
	var err error

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
	status           int32
	playerID         int32
	timer            float32
	totalGems        int32
	homing           int32
	enemiesAlive     int32
	enemiesKilled    int32
	daggersHit       int32
	daggersFired     int32
	level2time       float32
	level3time       float32
	level4time       float32
	isReplay         bool
	deathType        int32
	notifyPlayerBest bool
	notifyAbove1000  bool
	deathScreenSent  bool
}

func (c *Client) Submit(submissionData *SubmissionData) error {
	err := c.sioClient.Emit(
		submitFuncName,
		submissionData.playerID,
		submissionData.timer,
		submissionData.totalGems,
		submissionData.homing,
		submissionData.enemiesAlive,
		submissionData.enemiesKilled,
		submissionData.daggersHit,
		submissionData.daggersFired,
		submissionData.level2time,
		submissionData.level3time,
		submissionData.level4time,
		submissionData.isReplay,
		submissionData.deathType,
		submissionData.notifyPlayerBest,
		submissionData.notifyAbove1000,
	)
	if err != nil {
		return fmt.Errorf("Submit: error submitting: %w", err)
	}
	return nil
}

func (c *Client) SubmitStatusUpdate(playerID int, status int) error {
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
	f, err := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer f.Close()

	log.SetOutput(f)
	log.Printf("%v\n", inputErr)
	c.status = StatusDisconnected
}

func (c *Client) disconnectHandler() {
	c.status = StatusDisconnected
}
