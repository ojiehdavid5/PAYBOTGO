package handler

import (
	"fmt"

	// "github.com/chuks/PAYBOTGO/config"
	// "github.com/chuks/PAYBOTGO/models"
	"github.com/gofiber/fiber/v2"
	"github.com/chuks/PAYBOTGO/mono"
)

type MonoInitRequest struct {
	StudentID uint   `json:"student_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	
}

func InitiateMonoHandler(c *fiber.Ctx) error {
	var req MonoInitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	// Check if required fields are present (StudentID, Name, Email)
	if req.StudentID == 0 || req.Name == "" || req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing required fields"})
	}

	ref := fmt.Sprintf("student_ref_%d", req.StudentID)
	result, err := mono.InitiateMonoLink(req.Name, req.Email, ref)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// session := models.MonoSession{
	// 	StudentID:  req.StudentID,
	// 	Reference:  result.Data.Meta.Ref,
	// 	MonoURL:    result.Data.MonoURL,
	// 	CustomerID: result.Data.CustomerID,
	// }

	// if err := config.DB.Create(&session).Error; err != nil {
	// 	return c.Status(500).JSON(fiber.Map{"error": "could not save session"})
	// }

	return c.JSON(fiber.Map{
		"link": result.Data.MonoURL,
	})
}
