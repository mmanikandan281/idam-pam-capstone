package server

import (
	"database/sql"

	"idam-pam-platform/internal/config"
	"idam-pam-platform/internal/encryption"
	"idam-pam-platform/internal/handlers"
	"idam-pam-platform/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func New(cfg *config.Config, db *sql.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowCredentials: true,
	}))

	// Initialize services
	encryptionSvc := encryption.NewService(cfg.AWSRegion, cfg.KMSKeyID)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	userHandler := handlers.NewUserHandler(db)
	secretHandler := handlers.NewSecretHandler(db, encryptionSvc)
	auditHandler := handlers.NewAuditHandler(db)

	// Routes
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes
	protected := api.Use(middleware.JWTAuth(cfg.JWTSecret))
	protected.Use(middleware.EnsureUser(db))

	// User routes
	users := protected.Group("/users")
	// Read endpoints available to all authenticated users
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)
	// Write endpoints restricted to admins
	usersAdmin := users.Group("")
	usersAdmin.Use(middleware.RequireAdmin(db))
	usersAdmin.Put("/:id", userHandler.UpdateUser)
	usersAdmin.Post("/:id/roles", userHandler.AssignRole)

	// Secret routes
	secrets := protected.Group("/secrets")
	secrets.Get("/", secretHandler.GetSecrets)
	secrets.Post("/", secretHandler.CreateSecret)
	secrets.Get("/:id", secretHandler.GetSecret)
	secrets.Delete("/:id", secretHandler.DeleteSecret)

	// Audit routes
	audit := protected.Group("/audit")
	audit.Get("/", auditHandler.GetAuditLogs)

	// TOTP routes
	totp := protected.Group("/totp")
	totp.Post("/enable", authHandler.EnableTOTP)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	return app
}
