package net

import (
	"bytes"
	"encoding/json"
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

// func submitGame(gr GameRecording) {
// 	if (config.OfflineMode) ||
// 		(!config.Submit.Stats && gr.ReplayPlayerID == 0) ||
// 		(!config.Submit.ReplayStats && gr.ReplayPlayerID != 0) ||
// 		(!config.Submit.NonDefaultSpawnsets && gr.SurvivalHash != v3survivalHash) {
// 		return
// 	}
// 	debug.Log("Submitting Game.")
// 	jsonValue, err := json.Marshal(gr)
// 	if err != nil {
// 		debug.Log(err)
// 		gameRecording.Reset()
// 		return
// 	}

// 	resp, err := http.Post(config.Host+"/api/v2/submit_game", "application/json", bytes.NewBuffer(jsonValue))
// 	if err != nil {
// 		lastGameURL = "Error submitting game to server."
// 		return
// 	}

// 	var result map[string]interface{}

// 	json.NewDecoder(resp.Body).Decode(&result)

// 	debug.Log(result)

// 	if v, ok := result["game_id"]; ok {
// 		lastGameURL = fmt.Sprintf("https://ddstats.com/games/%v", v)
// 		if config.AutoClipboardGame {
// 			clipboard.WriteAll(lastGameURL)
// 		}
// 		if sioVariables.status == sioStatusLoggedIn {
// 			if (config.Submit.Stats && gr.ReplayPlayerID == 0) || (config.Submit.ReplayStats && gr.ReplayPlayerID != 0) {
// 				if !(!config.Submit.NonDefaultSpawnsets && gr.SurvivalHash != v3survivalHash) {
// 					sioClient.Emit("game_submitted", result["game_id"], config.Discord.NotifyPlayerBest, config.Discord.NotifyAbove1000)
// 				}
// 			}
// 		}
// 	} else if v, ok := result["message"]; ok {
// 		lastGameURL = v.(string)
// 	} else {
// 		lastGameURL = "No response received from server."
// 	}
// }
