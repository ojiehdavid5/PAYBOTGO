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
