package handlers

import (
	"database/sql"
	"encoding/json"

	"idam-pam-platform/internal/auth"
	"idam-pam-platform/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	db        *sql.DB
	jwtSecret string
}

func NewAuthHandler(db *sql.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Hash password
	passwordHash := auth.HashPassword(req.Password)

	// Insert user
	var userID uuid.UUID
	err := h.db.QueryRow(`
		INSERT INTO users (username, email, password_hash) 
		VALUES ($1, $2, $3) 
		RETURNING id`,
		req.Username, req.Email, passwordHash,
	).Scan(&userID)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Username or email already exists"})
	}

	// Log the registration
	h.logAudit(c, &userID, "user.register", "users", &userID, nil)

	return c.JSON(fiber.Map{
		"message": "User registered successfully",
		"user_id": userID,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get user from database
	var user models.User
	err := h.db.QueryRow(`
		SELECT id, username, email, password_hash, totp_secret, is_active 
		FROM users WHERE username = $1`,
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.TOTPSecret, &user.IsActive)

	if err != nil {
		h.logAudit(c, nil, "auth.login.failed", "auth", nil, map[string]string{
			"username": req.Username,
			"reason":   "user_not_found",
		})
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Check if user is active
	if !user.IsActive {
		h.logAudit(c, &user.ID, "auth.login.failed", "auth", nil, map[string]string{
			"reason": "user_inactive",
		})
		return c.Status(401).JSON(fiber.Map{"error": "Account is deactivated"})
	}

	// Verify password
	if !auth.VerifyPassword(req.Password, user.PasswordHash) {
		h.logAudit(c, &user.ID, "auth.login.failed", "auth", nil, map[string]string{
			"reason": "invalid_password",
		})
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Check TOTP if enabled
	if user.TOTPSecret != nil && *user.TOTPSecret != "" {
		if req.TOTPCode == "" {
			return c.JSON(fiber.Map{
				"requires_totp": true,
				"message":       "TOTP code required",
			})
		}

		if !auth.ValidateTOTP(req.TOTPCode, *user.TOTPSecret) {
			h.logAudit(c, &user.ID, "auth.login.failed", "auth", nil, map[string]string{
				"reason": "invalid_totp",
			})
			return c.Status(401).JSON(fiber.Map{"error": "Invalid TOTP code"})
		}
	}

	// Generate JWT
	token, err := auth.GenerateJWT(user.ID, user.Username, h.jwtSecret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	// Log successful login
	h.logAudit(c, &user.ID, "auth.login.success", "auth", nil, nil)

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (h *AuthHandler) EnableTOTP(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	uid, _ := uuid.Parse(userID)

	// Generate TOTP secret
	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate TOTP secret"})
	}

	// Update user with TOTP secret
	_, err = h.db.Exec(`
		UPDATE users SET totp_secret = $1 WHERE id = $2`,
		secret, uid,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to enable TOTP"})
	}

	// Generate QR code
	username := c.Locals("username").(string)
	qrCode, err := auth.GenerateQRCode(secret, username, "IDAM-PAM Platform")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate QR code"})
	}

	// Log TOTP enablement
	h.logAudit(c, &uid, "totp.enable", "users", &uid, nil)

	return c.JSON(fiber.Map{
		"secret": secret,
		"qr_url": qrCode.URL(),
	})
}

func (h *AuthHandler) logAudit(c *fiber.Ctx, userID *uuid.UUID, action, resource string, resourceID *uuid.UUID, details interface{}) {
	detailsJSON, _ := json.Marshal(details)
	
	_, err := h.db.Exec(`
		INSERT INTO audit_logs (user_id, action, resource, resource_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, action, resource, resourceID, detailsJSON, c.IP(), c.Get("User-Agent"),
	)
	if err != nil {
		// Log error but don't fail the request
		println("Failed to log audit:", err.Error())
	}
}