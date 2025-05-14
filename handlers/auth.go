package handlers

import (
	"database/sql"
	"encoding/json"
	"ithelp/db"
	"ithelp/models"
	"ithelp/utils"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	log.Println("Starting Register handler")
	
	var input struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Email    string `json:"email"` // optional
		Password string `json:"password"`
	}

	log.Println("Parsing request body")
	if err := c.BodyParser(&input); err != nil {
		log.Printf("Body parsing failed: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	log.Printf("Request received - Name: %s, Phone: %s, Email: %s", input.Name, input.Phone, input.Email)

	// Validate required fields
	log.Println("Validating required fields")
	if input.Name == "" || input.Phone == "" || input.Password == "" {
		log.Println("Validation failed: Missing required fields")
		return fiber.NewError(fiber.StatusBadRequest, "Name, phone, and password are required")
	}

	log.Println("Connecting to database")
	database, err := db.Connect()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database connection error")
	}
	defer database.Close()
	log.Println("Database connection established")

	// Check if phone already exists
	log.Printf("Checking if phone %s already exists", input.Phone)
	var exists bool
	err = database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE phone = ?)", input.Phone).Scan(&exists)
	if err != nil {
		log.Printf("Database error when checking phone: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	if exists {
		log.Printf("Phone %s already registered", input.Phone)
		return fiber.NewError(fiber.StatusConflict, "Phone number already registered")
	}
	log.Println("Phone is available")

	// Check if email exists (if provided)
	if input.Email != "" {
		log.Printf("Checking if email %s already exists", input.Email)
		err = database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", input.Email).Scan(&exists)
		if err != nil {
			log.Printf("Database error when checking email: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Database error")
		}
		if exists {
			log.Printf("Email %s already registered", input.Email)
			return fiber.NewError(fiber.StatusConflict, "Email already registered")
		}
		log.Println("Email is available")
	} else {
		log.Println("No email provided, skipping email check")
	}

	log.Println("Hashing password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Password hashing error")
	}
	log.Println("Password hashed successfully")

	log.Println("Inserting new user into database")
	var query string
	var result sql.Result
	
	// Handle empty email by using NULL in database
	if input.Email == "" {
		log.Println("Empty email, inserting NULL for email field")
		query = `INSERT INTO users (name, phone, email, password_hash, role) VALUES (?, ?, NULL, ?, ?)`
		result, err = database.Exec(query, input.Name, input.Phone, string(hashedPassword), "user")
	} else {
		log.Println("Inserting user with email")
		query = `INSERT INTO users (name, phone, email, password_hash, role) VALUES (?, ?, ?, ?, ?)`
		result, err = database.Exec(query, input.Name, input.Phone, input.Email, string(hashedPassword), "user")
	}
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create user")
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("User %s with phone %s registered successfully. Rows affected: %d", input.Name, input.Phone, rowsAffected)

	log.Println("Register handler completed successfully")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
	})
}

func Login(c *fiber.Ctx) error {
    log.Println("Starting Login handler")
    startTime := time.Now()

    var input struct {
        Phone    string `json:"phone"`
        Password string `json:"password"`
    }

    log.Println("Parsing request body")
    if err := c.BodyParser(&input); err != nil {
        log.Printf("Body parsing failed: %v", err)
        return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
    }
    log.Printf("Login attempt for phone: %s", input.Phone)

    log.Println("Validating required fields")
    if input.Phone == "" || input.Password == "" {
        log.Println("Validation failed: Missing phone or password")
        return fiber.NewError(fiber.StatusBadRequest, "Phone and password are required")
    }

    log.Println("Connecting to database")
    database, err := db.Connect()
    if err != nil {
        log.Printf("Database error: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "Database error")
    }
    defer database.Close()
    log.Println("Database connection established")

    log.Printf("Querying user data for phone: %s", input.Phone)
    var user models.User
    query := `SELECT id, name, email, password_hash, role FROM users WHERE phone = ?`
    err = database.QueryRow(query, input.Phone).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Printf("No user found with phone: %s", input.Phone)
            return fiber.NewError(fiber.StatusUnauthorized, "Invalid phone or password")
        }
        log.Printf("Database error during user query: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "Database error")
    }
    log.Printf("User found: ID=%d, Name=%s, Role=%s, Email=%+v", user.ID, user.Name, user.Role, user.Email) // Log the email value

    log.Println("Verifying password")
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
        log.Printf("Password verification failed for user ID: %d", user.ID)
        return fiber.NewError(fiber.StatusUnauthorized, "Invalid phone or password")
    }
    log.Println("Password verified successfully")

    log.Println("Generating JWT token")
    token, err := utils.GenerateJWT(user.ID, user.Role)
    if err != nil {
        log.Printf("Token generation failed: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "Token generation failed")
    }
    log.Printf("JWT token generated successfully for user ID: %d", user.ID)

    elapsed := time.Since(startTime)
    log.Printf("Login handler completed successfully in %v", elapsed)
	return c.JSON(fiber.Map{
        "success": true,
        "token":   token,
        "user": fiber.Map{
            "id":    user.ID,
            "name":  user.Name,
            "email": user.Email,
            "phone": user.Phone,
            "role":  user.Role,
        },
    })
}




func Me(c *fiber.Ctx) error {
	log.Println("Starting Me handler")
	
	log.Println("Getting user from context")
	userJson := c.Locals("userJSON").(string)
	log.Printf("User JSON from context: %s", userJson)

	var user models.User
	log.Println("Unmarshaling user JSON")
	if err := json.Unmarshal([]byte(userJson), &user); err != nil {
		log.Printf("Failed to unmarshal user JSON: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Invalid user context")
	}
	log.Printf("User unmarshaled successfully: ID=%d, Name=%s, Role=%s", user.ID, user.Name, user.Role)

	log.Println("Me handler completed successfully")
	return c.JSON(fiber.Map{
		"success": true,
		"user":    user,
	})
}