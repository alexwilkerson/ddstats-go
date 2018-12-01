package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func getMotd() {

	jsonData := map[string]string{"version": version}
	jsonValue, _ := json.Marshal(jsonData)
	resp, err := http.Post("https://ddstats.com/api/get_motd", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		motd = "Error getting MOTD."
		return
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	if v, ok := result["motd"]; ok {
		motd = v.(string)
	} else {
		motd = "Error fetching MOTD."
	}
	if v, ok := result["valid_version"]; ok {
		validVersion = v.(bool)
	}
	if v, ok := result["update_available"]; ok {
		updateAvailable = v.(bool)
	}

}

type whatever struct {
	name string
}

func submitGame(gr GameRecording) {
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
	} else if v, ok := result["message"]; ok {
		lastGameURL = v.(string)
	} else {
		lastGameURL = "No response received from server."
	}
}
