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
	PlayerID            int32     `json:"playerID"`
	PlayerName          string    `json:"playerName"`
	Granularity         int32     `json:"granularity"`
	Timer               float32   `json:"inGameTimer"`
	TimerSlice          []float32 `json:"inGameTimerVector"`
	TotalGems           int32     `json:"gems"`
	TotalGemsSlice      []int32   `json:"gemsVector"`
	Level2time          float32   `json:"levelTwoTime"`
	Level3time          float32   `json:"levelThreeTime"`
	Level4time          float32   `json:"levelFourTime"`
	LeviDownTime        float32   `json:"leviDownTime"`
	OrbDownTime         float32   `json:"orbDownTime"`
	Homing              int32     `json:"homingDaggers"`
	HomingSlice         []int32   `json:"homingDaggersVector"`
	HomingMax           int32     `json:"homingDaggersMax"`
	HomingMaxTime       float32   `json:"homingDaggersMaxTime"`
	DaggersFired        int32     `json:"daggersFired"`
	DaggersFiredSlice   []int32   `json:"daggersFiredVector"`
	DaggersHit          int32     `json:"daggersHit"`
	DaggersHitSlice     []int32   `json:"daggersHitVector"`
	EnemiesAlive        int32     `json:"enemiesAlive"`
	EnemiesAliveSlice   []int32   `json:"enemiesAliveVector"`
	EnemiesAliveMax     int32     `json:"enemiesAliveMax"`
	EnemiesAliveMaxTime float32   `json:"enemiesAliveMaxTime"`
	EnemiesKilled       int32     `json:"enemiesKilled"`
	EnemiesKilledSlice  []int32   `json:"enemiesKilledVector"`
	DeathType           uint8     `json:"deathType"`
	ReplayPlayerID      int32     `json:"replayPlayerID"`
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
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("SubmitGame: invalid response %s", resp.Status)
	}

	var result map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("SubmitGame: error decoding json: %w", err)
	}

	v, ok := result["game_id"]
	if !ok {
		return 0, errors.New("SubmitGame: no game id found in response")
	}

	return int(v.(float64)), nil
}
