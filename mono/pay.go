package mono

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type DirectPayRequest struct {
	Amount      int    `json:"amount"`
	Account     string `json:"account"`
	CustomerID  string `json:"customer_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Reference   string `json:"reference"`
	Description string `json:"description"`
}

func InitiateDirectPay(req DirectPayRequest) (string, error) {
	payload := map[string]interface{}{
		"amount":      req.Amount,
		"type":        "onetime-debit",
		"method":      "account",
		"account":     req.Account,
		"description": req.Description,
		"reference":   req.Reference,
		"redirect_url": "https://mono.co", // or your own frontend URL
		"customer": map[string]interface{}{
			"email":   req.Email,
			"phone":   req.Phone,
			"address": "N/A",
			"identity": map[string]interface{}{
				"type":   "bvn",
				"number": "22110033445", // Required, even if fake in sandbox
			},
			"name": req.Name,
		},
		"meta": map[string]string{},
	}

	jsonData, _ := json.Marshal(payload)
	url := "https://api.withmono.com/v2/payments/initiate"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	request.Header.Set("mono-sec-key", os.Getenv("MONO_SECRET_KEY"))
	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// Optional: print debug
	fmt.Println("🔍 Mono DirectPay response:", result)

	if link, ok := result["payment_link"].(string); ok {
		return link, nil
	}

	return "", fmt.Errorf("payment link not found: %+v", result)
}
