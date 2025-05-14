package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"
 
 
	"ithelp/db"
	"ithelp/models"
	
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Common error messages
const (
	ErrAccessDenied      = "Access denied"
	ErrDBConnection      = "Database connection failed"
	ErrUserNotFound      = "User not found"
	ErrInvalidInput      = "Invalid input"
	ErrInvalidUserID     = "Invalid user ID format"
	ErrInvalidAuth       = "Invalid authentication"
	ErrUserDataParse     = "Failed to parse user data"
	ErrQueryFailed       = "Database query failed"
	ErrUpdateFailed      = "Update failed"
	ErrDeleteFailed      = "Delete failed"
	ErrCreateFailed      = "Create failed"
	ErrOnlyAdminCreate   = "Only admins can create users"
	ErrPasswordHash      = "Failed to hash password"
)

// Helper function to parse and validate requester
func getRequester(c *fiber.Ctx) (*models.User, error) {
	userJson, ok := c.Locals("user").(string)
	if !ok {
		return nil, errors.New(ErrInvalidAuth)
	}

	var requester models.User
	if err := json.Unmarshal([]byte(userJson), &requester); err != nil {
		return nil, errors.New(ErrUserDataParse)
	}
	return &requester, nil
}

// Helper function to get database connection with context
func getDBWithContext() (*sql.DB, context.Context, context.CancelFunc, error) {
	database, err := db.Connect()
	if err != nil {
		return nil, nil, nil, err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return database, ctx, cancel, nil
}

func ListUsers(c *fiber.Ctx) error {
	requester, err := getRequester(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if requester.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": ErrAccessDenied,
		})
	}

	database, ctx, cancel, err := getDBWithContext()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrDBConnection,
		})
	}
	defer cancel()
	defer database.Close()

	rows, err := database.QueryContext(ctx, `
		SELECT id, name, phone, email, role, 
		subscription_plan, subscription_start, subscription_end 
		FROM users
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrQueryFailed,
		})
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Phone, &u.Email, &u.Role,
			&u.SubscriptionPlan, &u.SubscriptionStart, &u.SubscriptionEnd,
		); err != nil {
			continue
		}
		users = append(users, u)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// handlers/user.go
func GetUser(c *fiber.Ctx) error {
    // Get user claims from JWT
    claims, ok := c.Locals("userClaims").(models.User)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid user claims",
        })
    }

    idParam := c.Params("id")
    requestedID, err := strconv.Atoi(idParam)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID",
        })
    }

    // Authorization check
    if claims.Role != "admin" && claims.ID != requestedID {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Access denied",
        })
    }

    // Database query
    database, err := db.Connect()
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Database error",
        })
    }
    defer database.Close()

    var user models.User
    err = database.QueryRow(`
        SELECT id, name, phone, email, role,
        subscription_plan, subscription_start, subscription_end
        FROM users WHERE id = ?
    `, requestedID).Scan(
        &user.ID, &user.Name, &user.Phone, &user.Email, &user.Role,
        &user.SubscriptionPlan, &user.SubscriptionStart, &user.SubscriptionEnd,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "User not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Query failed",
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "data": user, // Return full user object
    })
}

func UpdateUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	targetID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidUserID,
		})
	}

	requester, err := getRequester(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if requester.Role != "admin" && requester.ID != targetID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": ErrAccessDenied,
		})
	}

	var input struct {
		Name             string `json:"name"`
		SubscriptionPlan string `json:"subscription_plan"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidInput,
		})
	}

	database, ctx, cancel, err := getDBWithContext()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrDBConnection,
		})
	}
	defer cancel()
	defer database.Close()

	_, err = database.ExecContext(ctx, `
		UPDATE users 
		SET name = ?, subscription_plan = ? 
		WHERE id = ?
	`, input.Name, input.SubscriptionPlan, targetID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrUpdateFailed,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User updated successfully",
	})
}

func DeleteUser(c *fiber.Ctx) error {
	requester, err := getRequester(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if requester.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": ErrAccessDenied,
		})
	}

	idParam := c.Params("id")
	targetID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidUserID,
		})
	}

	database, ctx, cancel, err := getDBWithContext()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrDBConnection,
		})
	}
	defer cancel()
	defer database.Close()

	_, err = database.ExecContext(ctx, `
		DELETE FROM users 
		WHERE id = ?
	`, targetID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrDeleteFailed,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})
}

func CreateUser(c *fiber.Ctx) error {
	requester, err := getRequester(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if requester.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": ErrOnlyAdminCreate,
		})
	}

	var input models.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidInput,
		})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrPasswordHash,
		})
	}

	database, ctx, cancel, err := getDBWithContext()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrDBConnection,
		})
	}
	defer cancel()
	defer database.Close()

	_, err = database.ExecContext(ctx, `
		INSERT INTO users (
			name, phone, email, password_hash, role
		) VALUES (?, ?, ?, ?, ?)
	`, input.Name, input.Phone, input.Email, string(hash), input.Role)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrCreateFailed,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
	})
}