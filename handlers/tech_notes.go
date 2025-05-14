package handlers

import (
	"encoding/json"
	"ithelp/db"
	"ithelp/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func AddTechNote(c *fiber.Ctx) error {
	userJson := c.Locals("user").(string)
	var tech models.User
	json.Unmarshal([]byte(userJson), &tech)

	if tech.Role != "tech" {
		return fiber.NewError(fiber.StatusForbidden, "Only technicians allowed")
	}

	var input struct {
		RequestID int    `json:"request_id"`
		Note      string `json:"note"`
	}

	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	dbConn, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()

	query := `INSERT INTO tech_notes (request_id, technician_id, note) VALUES (?, ?, ?)`
	_, err = dbConn.Exec(query, input.RequestID, tech.ID, input.Note)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Insert failed")
	}

	return c.JSON(fiber.Map{"success": true, "message": "Note added"})
}

func ListNotesByRequest(c *fiber.Ctx) error {
	requestID, _ := strconv.Atoi(c.Params("id"))

	dbConn, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()

	rows, err := dbConn.Query(`SELECT id, request_id, technician_id, note, created_at FROM tech_notes WHERE request_id = ?`, requestID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Query error")
	}
	defer rows.Close()

	var notes []models.TechNote
	for rows.Next() {
		var n models.TechNote
		err := rows.Scan(&n.ID, &n.RequestID, &n.TechnicianID, &n.Note, &n.CreatedAt)
		if err != nil {
			continue
		}
		notes = append(notes, n)
	}

	return c.JSON(fiber.Map{"success": true, "data": notes})
}