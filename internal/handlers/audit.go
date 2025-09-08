package handlers

import (
	"database/sql"

	"idam-pam-platform/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuditHandler struct {
	db *sql.DB
}

func NewAuditHandler(db *sql.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

func (h *AuditHandler) GetAuditLogs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	userID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(userID)

	// If user is admin, return all logs; else, return self logs
	var rows *sql.Rows
	var err error
	var isAdmin int
	_ = h.db.QueryRow(`
		SELECT 1 FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.name = 'admin'
		LIMIT 1
	`, uid).Scan(&isAdmin)

	if isAdmin == 1 {
		rows, err = h.db.Query(`
			SELECT a.id, a.user_id, a.action, a.resource, a.resource_id, a.details,
			       a.ip_address, a.user_agent, a.created_at, u.username
			FROM audit_logs a
			LEFT JOIN users u ON a.user_id = u.id
			ORDER BY a.created_at DESC
			LIMIT $1 OFFSET $2`,
			limit, offset,
		)
	} else {
		rows, err = h.db.Query(`
			SELECT a.id, a.user_id, a.action, a.resource, a.resource_id, a.details,
			       a.ip_address, a.user_agent, a.created_at, u.username
			FROM audit_logs a
			LEFT JOIN users u ON a.user_id = u.id
			WHERE a.user_id = $1
			ORDER BY a.created_at DESC
			LIMIT $2 OFFSET $3`,
			uid, limit, offset,
		)
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch audit logs"})
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var log models.AuditLog
		var username sql.NullString
		if err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.Resource, &log.ResourceID,
			&log.Details, &log.IPAddress, &log.UserAgent, &log.CreatedAt, &username); err != nil {
			continue
		}

		logs = append(logs, map[string]interface{}{
			"id":          log.ID,
			"user_id":     log.UserID,
			"username":    username.String,
			"action":      log.Action,
			"resource":    log.Resource,
			"resource_id": log.ResourceID,
			"details":     log.Details,
			"ip_address":  log.IPAddress,
			"user_agent":  log.UserAgent,
			"created_at":  log.CreatedAt,
		})
	}

	h.logAudit(c, &uid, "audit.list", "audit", nil, map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	return c.JSON(logs)
}

func (h *AuditHandler) logAudit(c *fiber.Ctx, userID *uuid.UUID, action, resource string, resourceID *uuid.UUID, details interface{}) {
	if action == "audit.list" {
		return
	}

	h.db.Exec(`
		INSERT INTO audit_logs (user_id, action, resource, resource_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, action, resource, resourceID, details, c.IP(), c.Get("User-Agent"),
	)
}
