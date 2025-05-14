package handlers

import (
	"database/sql"
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
		Phone    string `json:"phone"`
		Email    string `json:"email"` // optional
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	// Validate required fields
	if input.Name == "" || input.Phone == "" || input.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name, phone, and password are required")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database connection error")
	}
	defer database.Close()

	// Check if phone already exists
	var exists bool
	err = database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE phone = ?)", input.Phone).Scan(&exists)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if exists {
		return fiber.NewError(fiber.StatusConflict, "Phone number already registered")
	}

	// Check if email exists (if provided)
	if input.Email != "" {
		err = database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", input.Email).Scan(&exists)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Database error")
		}
		if exists {
			return fiber.NewError(fiber.StatusConflict, "Email already registered")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Password hashing error")
	}

	query := `INSERT INTO users (name, phone, email, password_hash, role) VALUES (?, ?, ?, ?, ?)`
	_, err = database.Exec(query, input.Name, input.Phone, input.Email, string(hashedPassword), "user")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create user")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
	})
}

func Login(c *fiber.Ctx) error {
	var input struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	if input.Phone == "" || input.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Phone and password are required")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer database.Close()

	var user models.User
	query := `SELECT id, name, email, password_hash, role FROM users WHERE phone = ?`
	err = database.QueryRow(query, input.Phone).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid phone or password")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid phone or password")
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