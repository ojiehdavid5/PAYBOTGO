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

	link, ok := result["mono_url"].(string)
	if !ok {
		// Optional: log full response for debugging
		fmt.Printf("Mono response: %+v\n", result)
		return "", fmt.Errorf("link not found in Mono response")
	}

	return link, nil
}
