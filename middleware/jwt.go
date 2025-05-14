// middleware/jwt.go
package middleware

import (
 
	"github.com/gofiber/fiber/v2"
	"strings"
	"ithelp/utils"
 
)
 
func JWTMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return fiber.NewError(fiber.StatusUnauthorized, "Missing authorization header")
        }

        // Extract token from "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return fiber.NewError(fiber.StatusUnauthorized, "Invalid authorization format")
        }

        token := parts[1]
        user, err := utils.ParseJWT(token)
        if err != nil {
            return fiber.NewError(fiber.StatusUnauthorized, "Invalid token: "+err.Error())
        }

        // Store the parsed user claims with proper type
        c.Locals("user", user) // Store the actual models.User struct
        return c.Next()
    }
}