package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	// "github.com/chuks/PAYBOTGO/config"
	// "github.com/chuks/PAYBOTGO/models"
	"github.com/chuks/PAYBOTGO/config"
	"github.com/chuks/PAYBOTGO/models"
	"github.com/chuks/PAYBOTGO/mono"
	"github.com/gofiber/fiber/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MonoInitRequest struct {
	StudentID uint   `json:"student_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

// func InitiateMonoHandler(c *fiber.Ctx) error {
// 	var req MonoInitRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
// 	}

// 	// Check if required fields are present (StudentID, Name, Email)
// 	if req.StudentID == 0 || req.Name == "" || req.Email == "" {
// 		return c.Status(400).JSON(fiber.Map{"error": "missing required fields"})
// 	}

// 	ref := fmt.Sprintf("student_ref_%d", req.StudentID)
// 	result, err := mono.InitiateMonoLink(req.Name, req.Email, ref)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	session := models.MonoSession{
// 		StudentID:  req.StudentID,
// 		Reference:  result.Data.Meta.Ref,
// 		MonoURL:    result.Data.MonoURL,
// 		CustomerID: result.Data.CustomerID,
// 	}

// 	if err := config.DB.Create(&session).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "could not save session"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"link": result.Data.MonoURL,
// 	})
// }



func InitiateMonoHandler(c *fiber.Ctx) error {
	var req MonoInitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.StudentID == 0 || req.Name == "" || req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing required fields"})
	}

	ref := fmt.Sprintf("student_ref_%d", req.StudentID)

	// Check if session already exists
	var existing models.MonoSession
	if err := config.DB.Where("reference = ?", ref).First(&existing).Error; err == nil {
		// Return existing link
		return c.JSON(fiber.Map{
			"link": existing.MonoURL,
		})
	}

	result, err := mono.InitiateMonoLink(req.Name, req.Email, ref)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	newSession := models.MonoSession{
		StudentID:  req.StudentID,
		Reference:  result.Data.Meta.Ref,
		MonoURL:    result.Data.MonoURL,
		CustomerID: result.Data.CustomerID,
	}

	if err := config.DB.Create(&newSession).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "could not save session"})
	}

	return c.JSON(fiber.Map{
		"link": result.Data.MonoURL,
	})
}


func HandleBalanceCheck(bot *tgbotapi.BotAPI, chatID int64) {
	go func() {
		var student models.Student
		dbResult := config.DB.Where("telegram_id = ?", chatID).First(&student)
		if dbResult.Error != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❗ No registered user found. Use /register first."))
			return
		}

		var session models.MonoSession
		if err := config.DB.Where("student_id = ?", student.ID).First(&session).Error; err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❗ Account not linked yet. Use /link_account first."))
			return
		}

		if session.AccountID == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Mono account not available yet. Please wait for confirmation."))
			return
		}

		url := fmt.Sprintf("https://api.withmono.com/v2/accounts/%s/balance", session.AccountID)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("mono-sec-key", os.Getenv("MONO_SECRET_KEY"))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Failed to check balance."))
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
json.NewDecoder(resp.Body).Decode(&result)

if data, ok := result["data"].(map[string]interface{}); ok {
	if balance, ok := data["balance"].(float64); ok {
		msg := fmt.Sprintf("💰 Your account balance is: ₦%.2f", balance/100)
		bot.Send(tgbotapi.NewMessage(chatID, msg))
		return
	}
}

bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Could not retrieve balance."))
	}()
}
