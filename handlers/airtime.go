package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)
// SendAirtime sends airtime to the specified phone number using the VTU API.

func SendAirtime(phone, amount string) error {
	payload := map[string]interface{}{
		"network":       "1", // Fixed for MTN
		"amount":        amount,
		"mobile_number": phone,
		"Ported_number": true,
		"airtime_type":  "VTU",
	}

	jsonData, _ := json.Marshal(payload)
	fmt.Println("🔍 VTU payload:", string(jsonData))

	req, err := http.NewRequest("POST", "https://mtsvtu.com/api/topup/", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	apiToken := os.Getenv("VTU_API_TOKEN")
	if apiToken == "" {
		return fmt.Errorf("VTU API token not set in environment")
	}

	// VTU_API_TOKEN="5051f0e5e0787cb41dbebe9d2793684892954b65"
	req.Header.Set("Authorization", "Token  "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s", string(bodyBytes))
	}

	return nil
}
