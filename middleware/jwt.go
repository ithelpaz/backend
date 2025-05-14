package middleware

import (
	"ithelp/utils"
	"github.com/gofiber/fiber/v2"
	"encoding/json"
)

func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing token")
		}

		user, err := utils.ParseJWT(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}

		userJSON, _ := json.Marshal(user)
		c.Locals("user", string(userJSON))
		return c.Next()
	}
}