package paystack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func CreatePaymentLink(email string, amount int) (string, error) {
	type Request struct {
		Email  string `json:"email"`
		Amount int    `json:"amount"` // in kobo (100 NGN = 10000)
	}
	type Response struct {
		Status bool `json:"status"`
		Data   struct {
			AuthorizationURL string `json:"authorization_url"`
		} `json:"data"`
	}

	body := Request{Email: email, Amount: amount}

	payload, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.paystack.co/transaction/initialize", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("PAYSTACK_SECRET_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res Response
	json.NewDecoder(resp.Body).Decode(&res)

	if !res.Status {
		return "", fmt.Errorf("failed to initialize payment")
	}

	return res.Data.AuthorizationURL, nil
}
