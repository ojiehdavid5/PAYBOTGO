package mono


import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Customer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Meta struct {
	Ref string `json:"ref"`
}

type RequestBody struct {
	Customer    Customer `json:"customer"`
	Meta        Meta     `json:"meta"`
	Scope       string   `json:"scope"`
	RedirectURL string   `json:"redirect_url"`
}

type MonoResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		MonoURL    string `json:"mono_url"`
		CustomerID string `json:"customer"`
		Meta       Meta   `json:"meta"`
	} `json:"data"`
}

func InitiateMonoLink(name, email, ref string) (*MonoResponse, error) {
	reqBody := RequestBody{
		Customer: Customer{
			Name:  name,
			Email: email,
		},
		Meta: Meta{Ref: ref},
		Scope:       "auth",
		RedirectURL: "https://t.me/Studentpay_bot",
	}

	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "https://api.withmono.com/v2/accounts/initiate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("mono-sec-key", os.Getenv("MONO_SECRET_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result MonoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "successful" {
		return nil, fmt.Errorf("mono failed: %s", result.Message)
	}

	return &result, nil
}
