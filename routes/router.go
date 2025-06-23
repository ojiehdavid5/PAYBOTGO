package routes

import (
	handler "github.com/chuks/PAYBOTGO/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {

	api := app.Group("/api")

	// Auth
	auth := api.Group("/auth")
	auth.Post("/login", handler.Login)
	auth.Post("/register", handler.Register)
	auth.Post("/verify", handler.VerifyOTP)
}
