package middleware

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// RequireAdmin allows the request to continue only if the current user has the 'admin' role.
func RequireAdmin(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDVal := c.Locals("userID")
		if userIDVal == nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}

		userID := userIDVal.(string)
		var exists int
		err := db.QueryRow(`
			SELECT 1
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = $1 AND r.name = 'admin'
			LIMIT 1
		`, userID).Scan(&exists)
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}

		return c.Next()
	}
}


