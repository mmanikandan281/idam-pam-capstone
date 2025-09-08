package main

import (
	"log"
	

	"idam-pam-platform/internal/config"
	"idam-pam-platform/internal/database"
	"idam-pam-platform/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Start server
	srv := server.New(cfg, db)
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(srv.Listen(":" + cfg.Port))
}

// roles : psql -h localhost -p 5432 -U postgres -d idam_pam -c "SELECT id, name FROM roles;"
//  7XVM2OYSH7EJVGIDJDB73TWFM7LKCHPC : Mkm
// 2PAZXCEQMUQKYKWAMIJHAF3E3HCKHJPU : admin