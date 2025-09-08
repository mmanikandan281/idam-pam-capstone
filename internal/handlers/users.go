package handlers

import (
	"database/sql"
	"encoding/json"

	"idam-pam-platform/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	rows, err := h.db.Query(`
		SELECT u.id, u.username, u.email, u.is_active, u.created_at, u.updated_at
		FROM users u
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	// Log the action
	userID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(userID)
	h.logAudit(c, &uid, "users.list", "users", nil, nil)

	return c.JSON(users)
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var user models.User
	err = h.db.QueryRow(`
		SELECT id, username, email, is_active, created_at, updated_at
		FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Get user roles
	roleRows, err := h.db.Query(`
		SELECT r.id, r.name, r.description
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1`,
		userID,
	)
	if err == nil {
		defer roleRows.Close()
		for roleRows.Next() {
			var role models.Role
			roleRows.Scan(&role.ID, &role.Name, &role.Description)
			user.Roles = append(user.Roles, role)
		}
	}

	// Log the action
	currentUserID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(currentUserID)
	h.logAudit(c, &uid, "users.read", "users", &userID, nil)

	return c.JSON(user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Update user
	_, err = h.db.Exec(`
		UPDATE users 
		SET is_active = COALESCE($2, is_active), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`,
		userID, updates["is_active"],
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
	}

	// Log the action
	currentUserID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(currentUserID)
	h.logAudit(c, &uid, "users.update", "users", &userID, updates)

	return c.JSON(fiber.Map{"message": "User updated successfully"})
}

func (h *UserHandler) AssignRole(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Params("id"))
	var req struct {
		RoleID uuid.UUID `json:"role_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	_, err := h.db.Exec(`
		INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING`,
		userID, req.RoleID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign role"})
	}

	// Log the action
	currentUserID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(currentUserID)
	h.logAudit(c, &uid, "users.assign_role", "users", &userID, map[string]interface{}{
		"role_id": req.RoleID,
	})

	return c.JSON(fiber.Map{"message": "Role assigned successfully"})
}

func (h *UserHandler) logAudit(c *fiber.Ctx, userID *uuid.UUID, action, resource string, resourceID *uuid.UUID, details interface{}) {
	detailsJSON, _ := json.Marshal(details)
	
	h.db.Exec(`
		INSERT INTO audit_logs (user_id, action, resource, resource_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, action, resource, resourceID, detailsJSON, c.IP(), c.Get("User-Agent"),
	)
}