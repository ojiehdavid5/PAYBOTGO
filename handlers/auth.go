package handler

import (
	"fmt"

	"github.com/chuks/PAYBOTGO/config"
	"github.com/chuks/PAYBOTGO/models"
	"github.com/chuks/PAYBOTGO/utils"
	"github.com/gofiber/fiber/v2"
)

type authRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type otpRequest struct {
		FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email    string `json:"email"`
	Password string `json:"password"`

	OTP      string `json:"otp"` // Corrected field name
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
		Email:    req.Email,
		Password: req.Password,
		FirstName: req.FirstName,
		LastName: req.LastName,
	}
	res := config.DB.Where("email = ?", user.Email).First(&models.Student{})
	if res.Error == nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "user already exists",
		})
	}

	otp, err := utils.SendOTP(user.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	body := fmt.Sprintf("Hello %s,\n\nYour OTP is: %s\nIt expires in 5 minutes.", req.Email, otp)
	err = utils.SendEmail(req.Email, "Verify your account", body)
	fmt.Println(err)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to send OTP"})
	}

	return c.JSON(fiber.Map{"message": "OTP sent to email"})

}
func VerifyOTP(c *fiber.Ctx) error {
	var req otpRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	fmt.Println("ok")

	valid, err := utils.VerifyOTP(req.Email, req.OTP) // Pass the actual OTP value
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if !valid {
		return c.Status(400).JSON(fiber.Map{
			"message": "invalid OTP",
		})
	}

	user := models.Student{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:    req.Email,
		Password: utils.GeneratePassword(req.Password),
	}
fmt.Println("ok")
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
	fmt.Println("ok")

	return c.JSON(fiber.Map{
		"token": token,
	})
}
