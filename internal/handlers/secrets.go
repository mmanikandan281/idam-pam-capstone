package handlers

import (
	"database/sql"
	"encoding/json"

	"idam-pam-platform/internal/encryption"
	"idam-pam-platform/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SecretHandler struct {
	db            *sql.DB
	encryptionSvc *encryption.Service
}

func NewSecretHandler(db *sql.DB, encryptionSvc *encryption.Service) *SecretHandler {
	return &SecretHandler{
		db:            db,
		encryptionSvc: encryptionSvc,
	}
}

func (h *SecretHandler) CreateSecret(c *fiber.Ctx) error {
	var req models.CreateSecretRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	userID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(userID)

	encryptedData, err := h.encryptionSvc.Encrypt(req.Data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to encrypt secret"})
	}

	var secretID uuid.UUID
	err = h.db.QueryRow(`
		INSERT INTO secrets (name, description, encrypted_data, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		req.Name, req.Description, encryptedData, uid,
	).Scan(&secretID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create secret"})
	}

	h.logAudit(c, &uid, "secrets.create", "secrets", &secretID, map[string]interface{}{
		"name": req.Name,
	})

	return c.JSON(fiber.Map{
		"id":      secretID,
		"message": "Secret created successfully",
	})
}

func (h *SecretHandler) GetSecrets(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(userID)

	rows, err := h.db.Query(`
		SELECT s.id, s.name, s.description, s.created_by, s.created_at, s.updated_at,
		       u.username as created_by_username
		FROM secrets s
		JOIN users u ON s.created_by = u.id
		WHERE s.created_by = $1
		ORDER BY s.created_at DESC
	`, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch secrets"})
	}
	defer rows.Close()

	var secrets []map[string]interface{}
	for rows.Next() {
		var secret models.Secret
		var createdByUsername string
		if err := rows.Scan(&secret.ID, &secret.Name, &secret.Description, &secret.CreatedBy,
			&secret.CreatedAt, &secret.UpdatedAt, &createdByUsername); err != nil {
			continue
		}

		secrets = append(secrets, map[string]interface{}{
			"id":                  secret.ID,
			"name":                secret.Name,
			"description":         secret.Description,
			"created_by":          secret.CreatedBy,
			"created_by_username": createdByUsername,
			"created_at":          secret.CreatedAt,
			"updated_at":          secret.UpdatedAt,
		})
	}

	h.logAudit(c, &uid, "secrets.list", "secrets", nil, nil)

	return c.JSON(secrets)
}

func (h *SecretHandler) GetSecret(c *fiber.Ctx) error {
	id := c.Params("id")
	secretID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid secret ID"})
	}

	userID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(userID)

	var secret models.Secret
	err = h.db.QueryRow(`
		SELECT id, name, description, encrypted_data, created_by, created_at, updated_at
		FROM secrets 
		WHERE id = $1 AND created_by = $2`,
		secretID, uid,
	).Scan(&secret.ID, &secret.Name, &secret.Description, &secret.EncryptedData,
		&secret.CreatedBy, &secret.CreatedAt, &secret.UpdatedAt)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Secret not found"})
	}

	decryptedData, err := h.encryptionSvc.Decrypt(secret.EncryptedData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decrypt secret"})
	}

	h.logAudit(c, &uid, "secrets.read", "secrets", &secretID, map[string]interface{}{
		"name": secret.Name,
	})

	return c.JSON(map[string]interface{}{
		"id":          secret.ID,
		"name":        secret.Name,
		"description": secret.Description,
		"data":        decryptedData,
		"created_by":  secret.CreatedBy,
		"created_at":  secret.CreatedAt,
		"updated_at":  secret.UpdatedAt,
	})
}

func (h *SecretHandler) DeleteSecret(c *fiber.Ctx) error {
	id := c.Params("id")
	secretID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid secret ID"})
	}

	userID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(userID)

	var secretName string
	err = h.db.QueryRow(`
		SELECT name FROM secrets WHERE id = $1 AND created_by = $2`,
		secretID, uid,
	).Scan(&secretName)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Secret not found"})
	}

	result, err := h.db.Exec(`DELETE FROM secrets WHERE id = $1 AND created_by = $2`, secretID, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete secret"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Secret not found"})
	}

	h.logAudit(c, &uid, "secrets.delete", "secrets", &secretID, map[string]interface{}{
		"name": secretName,
	})

	return c.JSON(fiber.Map{"message": "Secret deleted successfully"})
}

func (h *SecretHandler) logAudit(c *fiber.Ctx, userID *uuid.UUID, action, resource string, resourceID *uuid.UUID, details interface{}) {
	detailsJSON, _ := json.Marshal(details)

	h.db.Exec(`
		INSERT INTO audit_logs (user_id, action, resource, resource_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, action, resource, resourceID, detailsJSON, c.IP(), c.Get("User-Agent"),
	)
}
