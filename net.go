package main

import (
	"bytes"
	"encoding/json"
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

	motd = result["motd"].(string)
	validVersion = result["valid_version"].(bool)
	updateAvailable = result["update_available"].(bool)

}
