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
