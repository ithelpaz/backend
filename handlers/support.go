package handlers

import (
	"encoding/json"
	"ithelp/db"
	"ithelp/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func CreateSupportRequest(c *fiber.Ctx) error {
	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	defer database.Close()

	query := `INSERT INTO support_requests (user_id, title, description, category) VALUES (?, ?, ?, ?)`
	_, err = database.Exec(query, requester.ID, input.Title, input.Description, input.Category)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Insert error")
	}

	return c.JSON(fiber.Map{"success": true, "message": "Request created"})
}

func ListMySupportRequests(c *fiber.Ctx) error {
	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer database.Close()

	rows, err := database.Query("SELECT id, user_id, title, description, category, status, assigned_to, created_at, updated_at FROM support_requests WHERE user_id = ?", requester.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Query error")
	}
	defer rows.Close()

	var requests []models.SupportRequest
	for rows.Next() {
		var r models.SupportRequest
		err := rows.Scan(&r.ID, &r.UserID, &r.Title, &r.Description, &r.Category, &r.Status, &r.AssignedTo, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			continue
		}
		requests = append(requests, r)
	}

	return c.JSON(fiber.Map{"success": true, "data": requests})
}

func ListAllSupportRequests(c *fiber.Ctx) error {
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

	rows, err := database.Query("SELECT id, user_id, title, description, category, status, assigned_to, created_at, updated_at FROM support_requests")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Query error")
	}
	defer rows.Close()

	var requests []models.SupportRequest
	for rows.Next() {
		var r models.SupportRequest
		err := rows.Scan(&r.ID, &r.UserID, &r.Title, &r.Description, &r.Category, &r.Status, &r.AssignedTo, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			continue
		}
		requests = append(requests, r)
	}

	return c.JSON(fiber.Map{"success": true, "data": requests})
}

func UpdateSupportRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	requestID, _ := strconv.Atoi(idParam)

	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	if requester.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	var input struct {
		Status     string `json:"status"`
		AssignedTo *int   `json:"assigned_to"`
	}

	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer database.Close()

	query := `UPDATE support_requests SET status = ?, assigned_to = ? WHERE id = ?`
	_, err = database.Exec(query, input.Status, input.AssignedTo, requestID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Update failed")
	}

	return c.JSON(fiber.Map{"success": true, "message": "Request updated"})
}

func ListAssignedSupportRequests(c *fiber.Ctx) error {
	userJson := c.Locals("user").(string)
	var tech models.User
	json.Unmarshal([]byte(userJson), &tech)

	if tech.Role != "tech" {
		return fiber.NewError(fiber.StatusForbidden, "Only technicians allowed")
	}

	database, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer database.Close()

	rows, err := database.Query("SELECT id, user_id, title, description, category, status, assigned_to, created_at, updated_at FROM support_requests WHERE assigned_to = ?", tech.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Query error")
	}
	defer rows.Close()

	var requests []models.SupportRequest
	for rows.Next() {
		var r models.SupportRequest
		err := rows.Scan(&r.ID, &r.UserID, &r.Title, &r.Description, &r.Category, &r.Status, &r.AssignedTo, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			continue
		}
		requests = append(requests, r)
	}

	return c.JSON(fiber.Map{"success": true, "data": requests})
}