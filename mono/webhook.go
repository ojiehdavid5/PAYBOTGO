package mono

import (
	"fmt"

	"github.com/chuks/PAYBOTGO/config"
	"github.com/chuks/PAYBOTGO/models"
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
		var session models.MonoSession
		// Lookup MonoSession using CustomerID
		result := config.DB.Where("customer_id = ?", customerID).First(&session)
		if result.Error != nil {
			return c.Status(404).JSON(fiber.Map{"error": "no matching session found"})
		}

		// Update AccountID in that session
		session.AccountID = accountID
		if err := config.DB.Save(&session).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update session"})
		}

		fmt.Println("✅ MonoSession updated with AccountID")
		return c.SendStatus(200)
	}

	return c.SendStatus(204) // ignored
}
