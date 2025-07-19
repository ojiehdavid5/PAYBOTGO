package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chuks/PAYBOTGO/config"
	handler "github.com/chuks/PAYBOTGO/handlers"
	"github.com/chuks/PAYBOTGO/models"
	"github.com/chuks/PAYBOTGO/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// UserSession represents a user's session state
type UserSession struct {
	Step     string
	FullName string
	Email    string
	Passkey  string

	Password string
	Otp      string

	AirtimePhone   string
	AirtimeAmount  string
	AirtimeNetwork string
	PayAmount string

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
	// case "/pay":
	// 	fmt.Println("Creating Paystack link for:", session.Email)
	// 	// Assume session has user email
	// 	link, err := paystack.CreatePaymentLink(session.Email, 50000) // 500 NGN
	// 	if err != nil {
	// 		bot.Send(tgbotapi.NewMessage(chatID, "Error creating payment link: "+err.Error()))
	// 		return true
	// 	}
	// 	msg := tgbotapi.NewMessage(chatID, "Click below to pay securely via Paystack 👇\n"+link)
	// 	bot.Send(msg)
	case "/link_account":
		go func() {
			var student models.Student
			result := config.DB.Where("telegram_id = ?", chatID).First(&student)
			if result.Error != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❗ No registered user found. Use /register first."))
				return
			}

			req := map[string]interface{}{
				"student_id": student.ID,
				"name":       student.FirstName + " " + student.LastName,
				"email":      student.Email,
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

			fmt.Println("🔍 Mono response:", res)

			if link, ok := res["link"]; ok {
				msg := fmt.Sprintf("🔗 Click to link your bank account via Mono:\n%s", link)
				bot.Send(tgbotapi.NewMessage(chatID, msg))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Mono account linking failed."))
			}
		}()
		return true
	case "/balance":
		handler.HandleBalanceCheck(bot, chatID)
		return true
	case "/receipt":
		go func() {
			var student models.Student
			dbResult := config.DB.Where("telegram_id = ?", chatID).First(&student)
			if dbResult.Error != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❗ No registered user found. Use /register first."))
				return
			}

			// Simulate a transaction
			to := student.Email
			amount := 100000 // ₦1,000.00 in kobo
			txnID := fmt.Sprintf("TXN%d", time.Now().Unix())

			// Generate PDF
			pdfPath, err := utils.GenerateReceiptPDF(to, amount, txnID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Failed to generate receipt: "+err.Error()))
				return
			}

			// Send PDF
			doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(pdfPath))
			doc.Caption = fmt.Sprintf("🧾 Receipt for ₦%.2f", float64(amount)/100)
			if _, err := bot.Send(doc); err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Error sending receipt: "+err.Error()))
			}
		}()
		return true
	case "/airtime":
		session.Step = "awaiting_airtime_phone"
		bot.Send(tgbotapi.NewMessage(chatID, "📞 Enter the phone number to recharge:"))
		return true

	case "/pay":
		session.Step = "awaiting_pay_amount"
		bot.Send(tgbotapi.NewMessage(chatID, "💵 Enter the amount you want to pay (in Naira):"))
		return true

	default:
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
	case "awaiting_airtime_phone":
		session.AirtimePhone = text
		session.Step = "awaiting_airtime_amount"
		bot.Send(tgbotapi.NewMessage(chatID, "💵 Enter the amount (e.g. 500):"))

	case "awaiting_airtime_amount":
		session.AirtimeAmount = text
		session.Step = "awaiting_airtime_network"
		bot.Send(tgbotapi.NewMessage(chatID, "📡 Enter the network (mtn, glo, airtel, 9mobile):"))

	case "awaiting_airtime_network":
		session.AirtimeNetwork = strings.ToLower(text)
		go func() {
			bot.Send(tgbotapi.NewMessage(chatID, "⏳ Processing airtime request..."))

			err := handler.SendAirtime(session.AirtimePhone, session.AirtimeAmount)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Failed to send airtime: "+err.Error()))
			} else {
				msg := fmt.Sprintf("✅ Airtime of ₦%s successfully sent to %s (%s)", session.AirtimeAmount, session.AirtimePhone, session.AirtimeNetwork)
				bot.Send(tgbotapi.NewMessage(chatID, msg))
			}
		}()
		// session.Step = "" // reset step
	case "awaiting_pay_amount":
		session.PayAmount = text
		go func() {
			var student models.Student
			if err := config.DB.Where("telegram_id = ?", chatID).First(&student).Error; err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❗ Student not found. Use /register first."))
				return
			}

			var monoSession models.MonoSession
			if err := config.DB.Where("student_id = ?", student.ID).First(&monoSession).Error; err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❗ You need to /link_account first."))
				return
			}

			amountInt, err := strconv.Atoi(session.PayAmount)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Invalid amount. Please enter a number."))
				return
			}

			payload := map[string]interface{}{
				"amount":       amountInt * 100, // kobo
				"type":         "onetime-debit",
				"method":       "account",
				"account":      monoSession.AccountID,
				"description":  "Student bill payment",
				"reference":    fmt.Sprintf("ref_%d", time.Now().Unix()),
				"redirect_url": "https://mono.co",
				"customer": map[string]interface{}{
					"email":   student.Email,
					"phone":   "08122334455",
					"address": "school hostel",
					"name":    student.FirstName + " " + student.LastName,
					"identity": map[string]string{
						"type":   "bvn",
						"number": "22110033445", // test only; replace in prod
					},
				},
				"meta": map[string]string{},
			}

			data, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", "https://api.withmono.com/v2/payments/initiate", bytes.NewBuffer(data))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("mono-sec-key", os.Getenv("MONO_SECRET_KEY"))

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Payment initiation failed."))
				return
			}
			defer resp.Body.Close()

			var res map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&res)

			if data, ok := res["data"].(map[string]interface{}); ok {
				if link, ok := data["checkout_url"].(string); ok {
					bot.Send(tgbotapi.NewMessage(chatID, "🔗 Click below to authorize the payment:\n"+link))
					return
				}
			}

			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Could not initiate payment."))
		}()
		session.Step = ""

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
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Login successful. Proceed to payment link your account with Mono using /link_account."))
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
func sendAirtime(phone, amount string) error {
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
