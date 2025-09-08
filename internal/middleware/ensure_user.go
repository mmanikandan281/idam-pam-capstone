package middleware

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// EnsureUser upserts the user based on JWT claims and attaches a fresh user row.
func EnsureUser(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDVal := c.Locals("userID")
		usernameVal := c.Locals("username")
		if userIDVal == nil || usernameVal == nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}

		userID := userIDVal.(string)
		username := usernameVal.(string)

		// Upsert minimal user profile. Email is optional; if absent, keep existing.
		_, _ = db.Exec(`
			INSERT INTO users (id, username, email, password_hash, is_active)
			VALUES ($1, $2, $3, '', true)
			ON CONFLICT (id) DO UPDATE
			SET username = EXCLUDED.username,
			    updated_at = CURRENT_TIMESTAMP
		`, userID, username, username+"@example.com")

		return c.Next()
	}
}
