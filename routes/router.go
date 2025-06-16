package routes

import ("github.com/gofiber/fiber/v2"
	"github.com/chuks/PAYBOTGO/handlers"


)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {

	api := app.Group("/api")


	// Auth
	auth := api.Group("/auth")
	auth.Post("/login", handler.Login)
	auth.Post("/register", handler.Register)
}
