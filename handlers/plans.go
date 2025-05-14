package handlers

import (
	"encoding/json"
	"fmt"
	"ithelp/db"
	"ithelp/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func logAction(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s\n", timestamp, message)
}

func ListPlans(c *fiber.Ctx) error {
	logAction("ListPlans: Starting to retrieve plan list.")
	dbConn, err := db.Connect()
	if err != nil {
		logAction(fmt.Sprintf("ListPlans: Database connection error: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()
	logAction("ListPlans: Successfully connected to the database.")

	rows, err := dbConn.Query(`SELECT id, name, price, remote_calls, onsite_calls FROM plans`)
	if err != nil {
		logAction(fmt.Sprintf("ListPlans: Query execution error: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "Query error")
	}
	defer rows.Close()
	logAction("ListPlans: Successfully executed the query to fetch plans.")

	var plans []models.Plan
	for rows.Next() {
		var p models.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.RemoteCalls, &p.OnsiteCalls); err != nil {
			logAction(fmt.Sprintf("ListPlans: Error scanning row: %v", err))
			continue
		}
		plans = append(plans, p)
	}
	logAction(fmt.Sprintf("ListPlans: Retrieved %d plans.", len(plans)))

	return c.JSON(fiber.Map{"success": true, "data": plans})
}

func CreatePlan(c *fiber.Ctx) error {
	logAction("CreatePlan: Starting to create a new plan.")
	userJson := c.Locals("userJSON").(string)
	var requester models.User
	err := json.Unmarshal([]byte(userJson), &requester)
	if err != nil {
		logAction(fmt.Sprintf("CreatePlan: Error unmarshalling user from locals: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
	logAction(fmt.Sprintf("CreatePlan: Request made by user with role: %s", requester.Role))

	if requester.Role != "admin" {
		logAction("CreatePlan: Unauthorized attempt to create a plan.")
		return fiber.NewError(fiber.StatusForbidden, "Only admins can create plans")
	}

	var input models.Plan
	if err := c.BodyParser(&input); err != nil {
		logAction(fmt.Sprintf("CreatePlan: Invalid input received: %v", err))
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}
	logAction(fmt.Sprintf("CreatePlan: Received input: %+v", input))

	dbConn, err := db.Connect()
	if err != nil {
		logAction(fmt.Sprintf("CreatePlan: Database connection error: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()
	logAction("CreatePlan: Successfully connected to the database.")

	_, err = dbConn.Exec(`INSERT INTO plans (name, price, remote_calls, onsite_calls) VALUES (?, ?, ?, ?)`,
		input.Name, input.Price, input.RemoteCalls, input.OnsiteCalls)
	if err != nil {
		logAction(fmt.Sprintf("CreatePlan: Insert query failed: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "Insert failed")
	}
	logAction(fmt.Sprintf("CreatePlan: Plan '%s' created successfully.", input.Name))

	return c.JSON(fiber.Map{"success": true, "message": "Plan created"})
}

func UpdatePlan(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logAction(fmt.Sprintf("UpdatePlan: Invalid plan ID format: %s", idStr))
		return fiber.NewError(fiber.StatusBadRequest, "Invalid plan ID")
	}
	logAction(fmt.Sprintf("UpdatePlan: Starting to update plan with ID: %d", id))

	userJson := c.Locals("userJSON").(string)
	var requester models.User
	err = json.Unmarshal([]byte(userJson), &requester)
	if err != nil {
		logAction(fmt.Sprintf("UpdatePlan: Error unmarshalling user from locals: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
	logAction(fmt.Sprintf("UpdatePlan: Request made by user with role: %s", requester.Role))

	if requester.Role != "admin" {
		logAction(fmt.Sprintf("UpdatePlan: Unauthorized attempt to update plan ID: %d", id))
		return fiber.NewError(fiber.StatusForbidden, "Only admins can update plans")
	}

	var input models.Plan
	if err := c.BodyParser(&input); err != nil {
		logAction(fmt.Sprintf("UpdatePlan: Invalid input received: %v", err))
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}
	logAction(fmt.Sprintf("UpdatePlan: Received input: %+v for plan ID: %d", input, id))

	dbConn, err := db.Connect()
	if err != nil {
		logAction(fmt.Sprintf("UpdatePlan: Database connection error: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()
	logAction("UpdatePlan: Successfully connected to the database.")

	_, err = dbConn.Exec(`UPDATE plans SET name=?, price=?, remote_calls=?, onsite_calls=? WHERE id=?`,
		input.Name, input.Price, input.RemoteCalls, input.OnsiteCalls, id)
	if err != nil {
		logAction(fmt.Sprintf("UpdatePlan: Update query failed for plan ID %d: %v", id, err))
		return fiber.NewError(fiber.StatusInternalServerError, "Update failed")
	}
	logAction(fmt.Sprintf("UpdatePlan: Plan with ID %d updated successfully.", id))

	return c.JSON(fiber.Map{"success": true, "message": "Plan updated"})
}

func DeletePlan(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logAction(fmt.Sprintf("DeletePlan: Invalid plan ID format: %s", idStr))
		return fiber.NewError(fiber.StatusBadRequest, "Invalid plan ID")
	}
	logAction(fmt.Sprintf("DeletePlan: Starting to delete plan with ID: %d", id))

	userJson := c.Locals("userJSON").(string)
	var requester models.User
	err = json.Unmarshal([]byte(userJson), &requester)
	if err != nil {
		logAction(fmt.Sprintf("DeletePlan: Error unmarshalling user from locals: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
	logAction(fmt.Sprintf("DeletePlan: Request made by user with role: %s", requester.Role))

	if requester.Role != "admin" {
		logAction(fmt.Sprintf("DeletePlan: Unauthorized attempt to delete plan ID: %d", id))
		return fiber.NewError(fiber.StatusForbidden, "Only admins can delete plans")
	}

	dbConn, err := db.Connect()
	if err != nil {
		logAction(fmt.Sprintf("DeletePlan: Database connection error: %v", err))
		return fiber.NewError(fiber.StatusInternalServerError, "DB error")
	}
	defer dbConn.Close()
	logAction("DeletePlan: Successfully connected to the database.")

	_, err = dbConn.Exec(`DELETE FROM plans WHERE id=?`, id)
	if err != nil {
		logAction(fmt.Sprintf("DeletePlan: Delete query failed for plan ID %d: %v", id, err))
		return fiber.NewError(fiber.StatusInternalServerError, "Delete failed")
	}
	logAction(fmt.Sprintf("DeletePlan: Plan with ID %d deleted successfully.", id))

	return c.JSON(fiber.Map{"success": true, "message": "Plan deleted"})
}