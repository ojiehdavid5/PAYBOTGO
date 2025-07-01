package mono

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func InitiateMonoAccountLink(telegramID int64) (string, error) {
	reqBody := map[string]interface{}{
		"data": map[string]interface{}{
			"type":      "one_time", // or "recurring"
			"reference": fmt.Sprintf("student_%d", telegramID),
		},
	}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "https://api.withmono.com/v2/accounts/initiate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("mono-sec-key", os.Getenv("MONO_PUBLIC_KEY")) // Make sure this is correct!

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data field not found in response")
	}

	link, ok := data["mono_url"].(string)
	if !ok {
		return "", fmt.Errorf("mono_url not found in data")
	}
	return link, nil
}
