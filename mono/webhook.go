package mono

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type MonoWebhookData struct {
	Event string `json:"event"`
	Data  struct {
		ID       string `json:"id"`       // Account ID
		Customer string `json:"customer"` // Customer ID
	} `json:"data"`
}

func HandleMonoWebhook(c *fiber.Ctx) error {
	var webhook MonoWebhookData

	if err := c.BodyParser(&webhook); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	// Only process when account is connected
	if webhook.Event == "mono.events.account_connected" {
		accountID := webhook.Data.ID
		customerID := webhook.Data.Customer

		fmt.Println("✅ Account linked!")
		fmt.Println("Account ID:", accountID)
		fmt.Println("Customer ID:", customerID)

		// 👉 Store to DB or perform any action here
		// For example, you might want to link the Mono account with the user in your database
		
		return c.SendStatus(200)
	}

	return c.SendStatus(204) // ignored
}
