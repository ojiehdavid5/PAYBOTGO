package bot
import(
		tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
"github.com/joho/godotenv"
	"os"
	"log"
	"fmt"

)
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

}