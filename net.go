package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atotto/clipboard"
)

func getMotd() {

	jsonData := map[string]string{"version": version}
	jsonValue, _ := json.Marshal(jsonData)
	resp, err := http.Post(config.Host+"/api/v2/client_connect", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		if config.GetMOTD {
			motd = "Error getting MOTD."
		}
		return
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	if v, ok := result["motd"]; ok {
		if config.GetMOTD {
			motd = v.(string)
		}
	} else {
		if config.GetMOTD {
			motd = "Error fetching MOTD."
		}
	}
	if v, ok := result["valid_version"]; ok {
		validVersion = v.(bool)
	}
	if v, ok := result["update_available"]; ok {
		if config.CheckForUpdates {
			updateAvailable = v.(bool)
		}
	}

}

func submitGame(gr GameRecording) {
	if (config.OfflineMode) ||
		(!config.Submit.Stats && gr.ReplayPlayerID == 0) ||
		(!config.Submit.ReplayStats && gr.ReplayPlayerID != 0) ||
		(!config.Submit.NonDefaultSpawnsets && gr.SurvivalHash != v3survivalHash) {
		return
	}
	debug.Log("Submitting Game.")
	jsonValue, err := json.Marshal(gr)
	if err != nil {
		debug.Log(err)
		gameRecording.Reset()
		return
	}

	resp, err := http.Post(config.Host+"/api/v2/submit_game", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		lastGameURL = "Error submitting game to server."
		return
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	debug.Log(result)

	if v, ok := result["game_id"]; ok {
		lastGameURL = fmt.Sprintf("https://ddstats.com/games/%v", v)
		if config.AutoClipboardGame {
			clipboard.WriteAll(lastGameURL)
		}
		if sioVariables.status == sioStatusLoggedIn {
			if (config.Submit.Stats && gr.ReplayPlayerID == 0) || (config.Submit.ReplayStats && gr.ReplayPlayerID != 0) {
				if !(!config.Submit.NonDefaultSpawnsets && gr.SurvivalHash != v3survivalHash) {
					sioClient.Emit("game_submitted", result["game_id"], config.Discord.NotifyPlayerBest, config.Discord.NotifyAbove1000)
				}
			}
		}
	} else if v, ok := result["message"]; ok {
		lastGameURL = v.(string)
	} else {
		lastGameURL = "No response received from server."
	}
}
