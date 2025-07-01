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
	"github.com/chuks/PAYBOTGO/paystack"
)

// UserSession represents a user's session state
type UserSession struct {
	Step     string
	FullName string
	Email    string
	Passkey  string

	Password string
	Otp      string
}

var userStates = make(map[int64]*UserSession)

func StartBot() {
	_ = godotenv.Load()
	token := os.Getenv("TELEGRAM_APITOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_APITOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
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

		session := getOrCreateSession(chatID)

		if handleCommand(bot, chatID, text, session) {
			continue
		}

		handleConversation(bot, chatID, text, session)
	}
}

func handleCommand(bot *tgbotapi.BotAPI, chatID int64, text string, session *UserSession) bool {
	switch strings.ToLower(text) {
	case "/start":
		bot.Send(tgbotapi.NewMessage(chatID, "Welcome! Use /register or /login."))
		return true
	case "/register":
		session.Step = "awaiting_name"
		bot.Send(tgbotapi.NewMessage(chatID, "What's your full name?"))
		return true
	case "/login":
		session.Step = "awaiting_login_email"
		bot.Send(tgbotapi.NewMessage(chatID, "Enter your email:"))
		return true
	case "/verify_otp", "/verity_otp":
		session.Step = "awaiting_otp"
		bot.Send(tgbotapi.NewMessage(chatID, "Enter the OTP sent to your email:"))
		return true
	case "/pay":
		fmt.Println("Creating Paystack link for:", session.Email)
		// Assume session has user email
		link, err := paystack.CreatePaymentLink(session.Email, 50000) // 500 NGN
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Error creating payment link: "+err.Error()))
			return true
		}
		msg := tgbotapi.NewMessage(chatID, "Click below to pay securely via Paystack 👇\n"+link)
		bot.Send(msg)
case "/link_account":
	go func() {
		if session.FullName == "" || session.Email == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "❗ Please register first using /register before linking your account."))
			return
		}

req := map[string]interface{}{
    "customer": map[string]string{
        "name":  session.FullName,
        "email": session.Email,
    },
    "meta": map[string]string{
        "ref": fmt.Sprintf("student_%d", chatID),
    },
    "scope":        "auth",
    "redirect_url": "https://mono.co",
}

		body, _ := json.Marshal(req)
		fmt.Println("🔍 Mono payload:", string(body))
		resp, err := http.Post("http://localhost:3000/api/mono/initiate", "application/json", bytes.NewBuffer(body))
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Failed to connect to Mono API"))
			return
		}
		defer resp.Body.Close()

		var res map[string]string
		json.NewDecoder(resp.Body).Decode(&res)

		if link, ok := res["link"]; ok {
			msg := fmt.Sprintf("🔗 Click to link your bank account via Mono:\n%s", link)
			bot.Send(tgbotapi.NewMessage(chatID, msg))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Mono account linking failed."))
		}
	}()
	return true

	default:
		return false
	}
return false
}

func handleConversation(bot *tgbotapi.BotAPI, chatID int64, text string, session *UserSession) {
	switch session.Step {
	case "awaiting_name":
		session.FullName = text
		session.Step = "awaiting_email"
		bot.Send(tgbotapi.NewMessage(chatID, "What's your email?"))
	case "awaiting_email":
		session.Email = text
		session.Step = "awaiting_Passkey"
		bot.Send(tgbotapi.NewMessage(chatID, "Enter your Passkey:"))
	case "awaiting_Passkey":
		session.Passkey = text
		session.Step = "awaiting_password"
		bot.Send(tgbotapi.NewMessage(chatID, "Enter your password:"))
	case "awaiting_password":
		session.Password = text
		sendRegistration(bot, chatID, session)
	case "awaiting_login_email":
		session.Email = text
		session.Step = "awaiting_login_password"
		bot.Send(tgbotapi.NewMessage(chatID, "Enter your password:"))
	case "awaiting_login_password":
		session.Password = text
		sendLogin(bot, chatID, session)
	case "awaiting_otp":
		session.Otp = text
		sendOTPVerification(bot, chatID, session)
	}
}

func getOrCreateSession(chatID int64) *UserSession {
	session, exists := userStates[chatID]
	if !exists {
		session = &UserSession{}
		userStates[chatID] = session
	}
	return session
}

func sendRegistration(bot *tgbotapi.BotAPI, chatID int64, session *UserSession) {
	first, last := splitName(session.FullName)
	payload := map[string]interface{}{
		"first_name":  first,
		"last_name":   last,
		"email":       session.Email,
		"password":    session.Password,
		"passkey":     session.Passkey,
		"telegram_id": chatID,
	}
	err := callAPI("/api/auth/register", payload)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Registration failed: "+err.Error()))
	} else {
		session.Step = "awaiting_otp"
		bot.Send(tgbotapi.NewMessage(chatID, "OTP SENT TO YOUR EMAIL. Use /verify_otp to continue."))
	}
}

func sendLogin(bot *tgbotapi.BotAPI, chatID int64, session *UserSession) {
	payload := map[string]interface{}{
		"email":    session.Email,
		"password": session.Password,
	}
	err := callAPI("/api/auth/login", payload)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Login failed: "+err.Error()))
	} else {
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Login successful. Proceed to payment /pay"))
	}
}

func sendOTPVerification(bot *tgbotapi.BotAPI, chatID int64, session *UserSession) {
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
		bot.Send(tgbotapi.NewMessage(chatID, "✅ OTP verified. You are now registered! let link you account with mono /link_account"))
	}
}

func callAPI(endpoint string, payload map[string]interface{}) error {
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:3000"+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("bad request (likely user exists or bad input)")
	}

	return nil
}

func splitName(fullName string) (string, string) {
	names := strings.Fields(fullName)
	if len(names) > 1 {
		return names[0], strings.Join(names[1:], " ")
	}
	return fullName, ""
}
