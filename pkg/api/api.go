package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var defaultTimeout = 10 * time.Second

type Client struct {
	client  *http.Client
	hostURL string
}

func New(hostURL string) (*Client, error) {
	_, err := url.Parse(hostURL)
	if err != nil {
		return nil, fmt.Errorf("New: could not parse host url: %w", err)
	}

	return &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		hostURL: hostURL,
	}, nil
}

func (c *Client) WithClient(client *http.Client) *Client {
	c.client = client
	return c
}

type GetConnectionResult struct {
	MOTD            string `json:"motd"`
	ValidVersion    bool   `json:"valid_version"`
	UpdateAvailable bool   `json:"update_available"`
}

func (c *Client) InitConnection(version string) (*GetConnectionResult, error) {
	jsonData := map[string]string{"version": version}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("InitConnection: error marshalling JSON: %w", err)
	}

	resp, err := http.Post(c.hostURL+"/api/v2/client_connect", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("InitConnection: error connecting to site: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("InitConnection: invalid response %s", resp.Status)
	}

	var result GetConnectionResult

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("InitConnection: error decoding response: %w", err)
	}

	return &result, nil
}

type SubmitGameInput struct {
	PlayerID            uint32    `json:"playerID"`
	PlayerName          string    `json:"playerName"`
	Granularity         uint32    `json:"granularity"`
	Timer               float32   `json:"inGameTimer"`
	TimerSlice          []float32 `json:"inGameTimerVector"`
	TotalGems           uint32    `json:"gems"`
	TotalGemsSlice      []uint32  `json:"gemsVector"`
	Level2time          float32   `json:"levelTwoTime"`
	Level3time          float32   `json:"levelThreeTime"`
	Level4time          float32   `json:"levelFourTime"`
	Homing              uint32    `json:"homingDaggers"`
	HomingSlice         []uint32  `json:"homingDaggersVector"`
	HomingMax           uint32    `json:"homingDaggersMax"`
	HomingMaxTime       float32   `json:"homingDaggersMaxTime"`
	DaggersFired        uint32    `json:"daggersFired"`
	DaggersFiredSlice   []uint32  `json:"daggersFiredVector"`
	DaggersHit          uint32    `json:"daggersHit"`
	DaggersHitSlice     []uint32  `json:"daggersHitVector"`
	EnemiesAlive        uint32    `json:"enemiesAlive"`
	EnemiesAliveSlice   []uint32  `json:"enemiesAliveVector"`
	EnemiesAliveMax     uint32    `json:"enemiesAliveMax"`
	EnemiesAliveMaxTime float32   `json:"enemiesAliveMaxTime"`
	EnemiesKilled       uint32    `json:"enemiesKilled"`
	EnemiesKilledSlice  []uint32  `json:"enemiesKilledVector"`
	DeathType           uint32    `json:"deathType"`
	ReplayPlayerID      uint32    `json:"replayPlayerID"`
	Version             string    `json:"version"`
	SurvivalHash        string    `json:"survivalHash"`
}

func (c *Client) SubmitGame(submitGameInput *SubmitGameInput) (int, error) {
	jsonValue, err := json.Marshal(submitGameInput)
	if err != nil {
		return 0, fmt.Errorf("SubmitGame: error marshalling GameRecording: %w", err)
	}

	resp, err := http.Post(c.hostURL+"/api/v2/submit_game", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return 0, fmt.Errorf("SubmitGame: error connecting to site: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("SubmitGame: invalid response %s", resp.Status)
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	v, ok := result["game_id"]
	if !ok {
		return 0, errors.New("SubmitGame: no game id found in response")
	}

	return v.(int), nil
}
