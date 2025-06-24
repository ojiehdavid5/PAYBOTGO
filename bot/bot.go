package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type UserSession struct {
	Step     string
	FullName string
	Email    string
	Password string
	Passkey  string
	Otp      string
}

var userStates = make(map[int64]*UserSession)

func StartBot() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_APITOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_APITOKEN environment variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text
		session, exists := userStates[chatID]

		if !exists && text == "/register" {
			userStates[chatID] = &UserSession{Step: "awaiting_name"}
			bot.Send(tgbotapi.NewMessage(chatID, "What's your full name?"))
			continue
		}

		if exists {
			switch session.Step {
			case "awaiting_name":
				session.FullName = text
				session.Step = "awaiting_email"
				bot.Send(tgbotapi.NewMessage(chatID, "What's your email?"))

			case "awaiting_email":
				session.Email = text
				session.Step = "awaiting_password"
				bot.Send(tgbotapi.NewMessage(chatID, "Enter your password:"))

			case "awaiting_password":
				session.Password = text
				session.Step = "awaiting_passkey"
				bot.Send(tgbotapi.NewMessage(chatID, "Set a 4–6 digit passkey (PIN) for payment confirmation:"))

			case "awaiting_passkey":
				session.Passkey = text
				session.Step = "awaiting_otp"
				bot.Send(tgbotapi.NewMessage(chatID, "OTP sent to your email. Please enter it using /verify_otp"))

			case "awaiting_otp":
				session.Otp = text

				first, last := splitName(session.FullName)

				payload := map[string]interface{}{
					"first_name":  first,
					"last_name":   last,
					"email":       session.Email,
					"password":    session.Password,
					"passkey":     session.Passkey,
					"telegram_id": chatID,
					"otp":         session.Otp,
				}

				err := callAPI("/api/auth/verify", payload)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "OTP verification failed: "+err.Error()))
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "✅ OTP verified successfully. You are now registered!"))
					delete(userStates, chatID) // clear session only after verification
				}

			case "awaiting_login_email":
				session.Email = text
				session.Step = "awaiting_login_password"
				bot.Send(tgbotapi.NewMessage(chatID, "Enter your password:"))

			case "awaiting_login_password":
				session.Password = text
				delete(userStates, chatID)

				payload := map[string]any{
					"email":    session.Email,
					"password": session.Password,
				}
				err := callAPI("/api/auth/login", payload)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Login failed."))
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "Login successful."))
				}
			}
			continue
		}

		// Handle commands
		switch text {
		case "/start":
			bot.Send(tgbotapi.NewMessage(chatID, "Welcome! Use /register or /login."))

		case "/register":
			bot.Send(tgbotapi.NewMessage(chatID, "Let's register you. What's your full name?"))
			userStates[chatID] = &UserSession{Step: "awaiting_name"}

		case "/login":
			bot.Send(tgbotapi.NewMessage(chatID, "Enter your email:"))
			userStates[chatID] = &UserSession{Step: "awaiting_login_email"}

		case "/verify_otp":
			if session != nil {
				session.Step = "awaiting_otp"
				bot.Send(tgbotapi.NewMessage(chatID, "Enter the OTP sent to your email:"))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "No pending verification. Please register first."))
			}

		default:
			bot.Send(tgbotapi.NewMessage(chatID, "Invalid command. Use /register or /login."))
		}
	}
}

func splitName(fullName string) (string, string) {
	names := strings.Fields(fullName)
	if len(names) > 1 {
		return names[0], strings.Join(names[1:], " ")
	}
	return fullName, ""
}

func callAPI(endpoint string, payload map[string]interface{}) error {
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:3000"+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("bad request (e.g. user already exists)")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response: %s", resp.Status)
	}
	return nil
}
