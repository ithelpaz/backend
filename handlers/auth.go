package handlers

import (
	"encoding/json"
	"ithelp/db"
	"ithelp/models"
	"ithelp/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database connection error")
	}
	defer database.Close()

	query := `INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)`
	_, err = database.Exec(query, input.Name, input.Email, string(hashedPassword), "user")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Email already exists or DB error")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
	})
}

func Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer database.Close()

	var user models.User
	query := `SELECT id, name, email, password_hash, role FROM users WHERE email = ?`
	err = database.QueryRow(query, input.Email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid email or password")
	}

	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Token generation failed")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"token":   token,
	})
}

func Me(c *fiber.Ctx) error {
	userJson := c.Locals("user").(string)

	var user models.User
	if err := json.Unmarshal([]byte(userJson), &user); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Invalid user context")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"user":    user,
	})
}