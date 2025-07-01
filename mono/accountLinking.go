package mono

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func InitiateMonoAccountLink(telegramID int64) (string, error) {
	requestBody := map[string]interface{}{
		"data": map[string]interface{}{
			"type":      "one_time", // or "recurring"
			"reference": fmt.Sprintf("student_%d", telegramID),
		},
	}

	jsonData, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", "https://api.withmono.com/v2/accounts/initiate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("mono-sec-key", "test_pk_fqdjxmqwhot8bx09b1qp")

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

	// Extract link
	if link, ok := result["link"].(string); ok {
		return link, nil
	}

	return "", fmt.Errorf("link not found in response")
}
