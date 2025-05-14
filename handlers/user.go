package handlers

import (

	"ithelp/db"
	"ithelp/models"
	"encoding/json"

	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ListUsers(c *fiber.Ctx) error {
	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	if requester.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer database.Close()

	rows, err := database.Query("SELECT id, name, email, role, subscription_plan, subscription_start, subscription_end FROM users")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Query error")
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.SubscriptionPlan, &u.SubscriptionStart, &u.SubscriptionEnd)
		if err != nil {
			continue
		}
		users = append(users, u)
	}

	return c.JSON(fiber.Map{"success": true, "data": users})
}

func GetUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	requestedID, _ := strconv.Atoi(idParam)

	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	if requester.Role != "admin" && requester.ID != requestedID {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer database.Close()
	
	var u models.User
	query := `SELECT id, name, email, role, subscription_plan, subscription_start, subscription_end FROM users WHERE id = ?`
	err = database.QueryRow(query, requestedID).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.SubscriptionPlan, &u.SubscriptionStart, &u.SubscriptionEnd)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	return c.JSON(fiber.Map{"success": true, "data": u})
}

func UpdateUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	targetID, _ := strconv.Atoi(idParam)

	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	if requester.Role != "admin" && requester.ID != targetID {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	var input struct {
		Name             string `json:"name"`
		SubscriptionPlan string `json:"subscription_plan"`
	}

	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer database.Close()

	query := `UPDATE users SET name = ?, subscription_plan = ? WHERE id = ?`
	_, err = database.Exec(query, input.Name, input.SubscriptionPlan, targetID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Update failed")
	}

	return c.JSON(fiber.Map{"success": true, "message": "User updated"})
}

func DeleteUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	targetID, _ := strconv.Atoi(idParam)

	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	if requester.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer database.Close()

	query := `DELETE FROM users WHERE id = ?`
	_, err = database.Exec(query, targetID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Delete failed")
	}

	return c.JSON(fiber.Map{"success": true, "message": "User deleted"})
}