package bot

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	// "github.com/chuks/PAYBOTGO/config"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

// UserSession represents the state of a user's registration process
type UserSession struct {
	Step     string
	FullName string
	Email    string
	Password string
	Otp      string
	Passkey string
}

// Global map to store user states
var userStates = make(map[int64]*UserSession)

func StartBot() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the bot token from the environment variable
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
				delete(userStates, chatID) // remove session

				case "awaiting_passkey":
				session.Passkey = text
				session.Step = "awaiting_otp"
				bot.Send(tgbotapi.NewMessage(chatID, "OTP sent to your email. Please enter it using /verify_otp"))

				first, last := splitName(session.FullName)
				payload := map[string]interface{}{
					"first_name":  first,
					"last_name":   last,
					"email":       session.Email,
					"password":    session.Password,
					"telegram_id": chatID,
				}
				err := callAPI("/api/auth/register", payload)

				fmt.Println(err)

				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, err.Error()))
					return
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "OTP SENT TO YOUR EMAIL verify /verify_otp"))
				}


				case "awaiting_otp":
				session.Otp = text
				delete(userStates, chatID) // Clear session after attempt
				
				// first, last := splitName(session.FullName)

				payload := map[string]interface{}{
					// "first_name":  first,
					// "last_name":   last,
					"email":       session.Email,
					"password":    session.Password,
					"telegram_id": chatID,
					"otp":         session.Otp,
				}
				fmt.Println(payload)

				err := callAPI("/api/auth/verify", payload)
				fmt.Println(err)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "OTP verification failed: "+err.Error()))
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "✅ OTP verified successfully. You are now registered!"))
				}


			case "awaiting_login_email":
				session.Email = text
				session.Step = "awaiting_login_password"
				bot.Send(tgbotapi.NewMessage(chatID, "Enter your password:"))
			case "awaiting_login_password":
				session.Password = text
				delete(userStates, chatID) // remove session

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
				fmt.Println(payload)
				bot.Send(tgbotapi.NewMessage(chatID, "Login submitted."))

			
			}

			continue

		}

		switch text {
		case "/start":
			bot.Send(tgbotapi.NewMessage(chatID, "Welcome! Use /register or /login."))

		case "/register":
			bot.Send(tgbotapi.NewMessage(chatID, "Send your details like: `FirstName|LastName|email@example.com|password`"))
			userStates[chatID] = &UserSession{Step: "awaiting_name"}

		case "/login":
			bot.Send(tgbotapi.NewMessage(chatID, "Send email"))
			userStates[chatID] = &UserSession{Step: "awaiting_login_email"}

		case "/verify_otp":
			bot.Send(tgbotapi.NewMessage(chatID, "Enter the OTP sent to your email:"))
			userStates[chatID] = &UserSession{Step: "awaiting_otp"}

		default:
			bot.Send(tgbotapi.NewMessage(chatID, "Invalid command."))

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

// Helper function to call API (you need to implement this)
func callAPI(endpoint string, payload map[string]interface{}) error {

	// Implement API call logic here
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:3000"+endpoint, "application/json", bytes.NewBuffer(jsonData))

	fmt.Println(err)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("user already exists")
	}

	return nil
}
