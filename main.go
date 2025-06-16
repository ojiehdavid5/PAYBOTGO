package main

import (
	"fmt"

	"github.com/chuks/PAYBOTGO/bot"
	"github.com/chuks/PAYBOTGO/config"
	"github.com/gofiber/fiber/v2"
	"github.com/chuks/PAYBOTGO/handlers"
)

func main() {
	config.ConnectDB()
		bot.StartBot()
	fmt.Println("Bot started")


		app := fiber.New()

	app.Post("/api/auth/register", handler.Register)

app.Listen(":3000")



	// fmt.Println("chuks on chains")
}
