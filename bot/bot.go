package bot
import(
		tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
"github.com/joho/godotenv"
	"os"
	"log"
	"fmt"
		// "github.com/gofiber/fiber/v2"
		// "github.com/chuks/PAYBOTGO/config"


)
func StartBot() {
		err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

// db, err := database.Connect()
// 	if err != nil {
// 		log.Fatalf("Error connecting to database: %v", err)
// 	}
// 	// fmt.Println(db)
// 	app := fiber.New()



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

        switch text {
        case "/start":
            bot.Send(tgbotapi.NewMessage(chatID, "Welcome! Use /register or /login."))

        case "/register":
            bot.Send(tgbotapi.NewMessage(chatID, "Send your details like: `FirstName|LastName|email@example.com|password`"))
			fmt.Println(update)

        case "/login":
            bot.Send(tgbotapi.NewMessage(chatID, "Send email and password like: `email@example.com|password`"))

        default:
            bot.Send(tgbotapi.NewMessage(chatID, "Invalid command."))
		}
	}
}	