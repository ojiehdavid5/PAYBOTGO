package handler

import (
	"github.com/chuks/PAYBOTGO/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/chuks/PAYBOTGO/config"
	"github.com/chuks/PAYBOTGO/models"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(c *fiber.Ctx) error {
	// Get the user session from the context

	var req authRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	user := models.Student{
	Email:        req.Email,
		Password: utils.GeneratePassword(req.Password),
	}
	res := config.DB.Create(&user)
	if res.Error != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": res.Error.Error(),
		})
	}
	return c.Status(201).JSON(fiber.Map{
		"message": "user created",
	})
}

func Login(c *fiber.Ctx) error {
 var req authRequest
 if err := c.BodyParser(&req); err != nil {
  return c.Status(400).JSON(fiber.Map{
   "message": err.Error(),
  })
 }
 var user models.Student
 res := config.DB.Where("email = ?", req.Email).First(&user)
 if res.Error != nil {
  return c.Status(400).JSON(fiber.Map{
   "message": "user not found",
  })
 }
 if !utils.ComparePassword(user.Password, req.Password) {
  return c.Status(400).JSON(fiber.Map{
   "message": "incorrect password",
  })
 }

 token, err := utils.GenerateToken(user.ID)
 if err != nil {
  return c.Status(500).JSON(fiber.Map{
   "message": err.Error(),
  })
 }
 return c.JSON(fiber.Map{
  "token": token,
 })
}
