package handlers

import (
	"encoding/json"
	"ithelp/db"
	"ithelp/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ListPlans(c *fiber.Ctx) error {
	dbConn, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()

	rows, err := dbConn.Query(`SELECT id, name, price, remote_calls, onsite_calls FROM plans`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Query error")
	}
	defer rows.Close()

	var plans []models.Plan
	for rows.Next() {
		var p models.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.RemoteCalls, &p.OnsiteCalls); err != nil {
			continue
		}
		plans = append(plans, p)
	}

	return c.JSON(fiber.Map{"success": true, "data": plans})
}

func CreatePlan(c *fiber.Ctx) error {
	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	if requester.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admins can create plans")
	}

	var input models.Plan
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	dbConn, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()

	_, err = dbConn.Exec(`INSERT INTO plans (name, price, remote_calls, onsite_calls) VALUES (?, ?, ?, ?)`,
		input.Name, input.Price, input.RemoteCalls, input.OnsiteCalls)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Insert failed")
	}

	return c.JSON(fiber.Map{"success": true, "message": "Plan created"})
}

func UpdatePlan(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	if requester.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admins can update plans")
	}

	var input models.Plan
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	dbConn, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()

	_, err = dbConn.Exec(`UPDATE plans SET name=?, price=?, remote_calls=?, onsite_calls=? WHERE id=?`,
		input.Name, input.Price, input.RemoteCalls, input.OnsiteCalls, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Update failed")
	}

	return c.JSON(fiber.Map{"success": true, "message": "Plan updated"})
}

func DeletePlan(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	userJson := c.Locals("user").(string)
	var requester models.User
	json.Unmarshal([]byte(userJson), &requester)

	if requester.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admins can delete plans")
	}

	dbConn, err := db.Connect()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()

	_, err = dbConn.Exec(`DELETE FROM plans WHERE id=?`, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Delete failed")
	}

	return c.JSON(fiber.Map{"success": true, "message": "Plan deleted"})
}