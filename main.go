package main

import (
	"fmt"

	"github.com/chuks/PAYBOTGO/bot"
	"github.com/chuks/PAYBOTGO/config"
	handler "github.com/chuks/PAYBOTGO/handlers"
	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()

	// Set up routes
	app.Post("/api/auth/register", handler.Register)
	app.Post("/api/auth/login", handler.Login)
	app.Post("/api/auth/verify", handler.VerifyOTP)
	app.Post("/paystack/webhook", func(c *fiber.Ctx) error {
	// Validate Paystack signature if needed
	var event map[string]interface{}
	if err := c.BodyParser(&event); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if event["event"] == "charge.success" {
		data := event["data"].(map[string]interface{})
		email := data["customer"].(map[string]interface{})["email"].(string)

		// ✅ Mark user as paid in DB or notify via Telegram bot
		fmt.Println("Payment successful for:", email)
	}
	return c.SendStatus(fiber.StatusOK)
})
app.Post("/api/mono/initiate", handler.InitiateMonoHandler)


	// Connect to the database
	config.ConnectDB()

	// Start the bot in a separate goroutine
	go func() {
		bot.StartBot()
		fmt.Println("Bot started")
	}()

	// Start the server
	fmt.Println("Server is running on port 3000")
	err := app.Listen(":3000")
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}

	// fmt.Println("chuks on chains")
}
