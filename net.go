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
	resp, err := http.Post("https://ddstats.com/api/get_motd", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		if config.getMOTD {
			motd = "Error getting MOTD."
		}
		return
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	if v, ok := result["motd"]; ok {
		if config.getMOTD {
			motd = v.(string)
		}
	} else {
		if config.getMOTD {
			motd = "Error fetching MOTD."
		}
	}
	if v, ok := result["valid_version"]; ok {
		validVersion = v.(bool)
	}
	if v, ok := result["update_available"]; ok {
		if config.checkForUpdates {
			updateAvailable = v.(bool)
		}
	}

}

func submitGame(gr GameRecording) {
	if (config.offlineMode) ||
		(!config.submit.stats && gr.ReplayPlayerID == 0) ||
		(!config.submit.replayStats && gr.ReplayPlayerID != 0) ||
		(!config.submit.nonDefaultSpawnsets && gr.SurvivalHash != v3survivalHash) {
		return
	}
	debug.Log("Submitting Game.")
	jsonValue, err := json.Marshal(gr)
	if err != nil {
		debug.Log(err)
		gameRecording.Reset()
		return
	}

	resp, err := http.Post("http://ddstats.com/api/submit_game", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		lastGameURL = "Error submitting game to server."
		return
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	debug.Log(result)

	if v, ok := result["game_id"]; ok {
		lastGameURL = fmt.Sprintf("https://ddstats.com/game_log/%v", v)
		if config.autoClipboardGame {
			clipboard.WriteAll(lastGameURL)
		}
		if sioVariables.status == sioStatusLoggedIn {
			if (config.stream.stats && gr.ReplayPlayerID == 0) || (config.stream.replayStats && gr.ReplayPlayerID != 0) {
				if !(!config.stream.nonDefaultSpawnsets && gr.SurvivalHash != v3survivalHash) {
					sioClient.Emit("game_submitted", result["game_id"], config.discord.notifyPlayerBest, config.discord.notifyAbove1000)
				}
			}
		}
	} else if v, ok := result["message"]; ok {
		lastGameURL = v.(string)
	} else {
		lastGameURL = "No response received from server."
	}
}
