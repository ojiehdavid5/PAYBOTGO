package bot
import(
		tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
"github.com/joho/godotenv"
	"os"
	"log"
	"fmt"
		// "github.com/gofiber/fiber/v2"
		// "github.com/chuks/PAYBOTGO/config"
		"strings"
	"bytes"
	"encoding/json"
	"net/http"


)
// UserSession represents the state of a user's registration process
type UserSession struct {
    Step     string
    FullName string
    Email    string
    Password string
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
        if update.Message == nil { continue }

        chatID := update.Message.Chat.ID
        text := update.Message.Text

		session,exists := userStates[chatID]

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

                first, last := splitName(session.FullName)
                payload := map[string]interface{}{
                    "first_name":  first,
                    "last_name":   last,
                    "email":       session.Email,
                    "password":    session.Password,
                    "telegram_id": chatID,
                }
                callAPI("/api/auth/register", payload)
				fmt.Println(payload)
                bot.Send(tgbotapi.NewMessage(chatID, "Registration submitted."))
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
            bot.Send(tgbotapi.NewMessage(chatID, "Send email and password like: `email@example.com|password`"))

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
func callAPI(endpoint string, payload map[string]interface{}) {
    // Implement API call logic here
	jsonData, _ := json.Marshal(payload)
	http.Post("http://localhost:3000"+endpoint, "application/json", bytes.NewBuffer(jsonData))
}