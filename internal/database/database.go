package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Init(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			totp_secret VARCHAR(255),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS permissions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) UNIQUE NOT NULL,
			resource VARCHAR(255) NOT NULL,
			action VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS user_roles (
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
			assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, role_id)
		);`,

		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
			permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
			assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (role_id, permission_id)
		);`,

		`CREATE TABLE IF NOT EXISTS secrets (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			encrypted_data TEXT NOT NULL,
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID REFERENCES users(id),
			action VARCHAR(255) NOT NULL,
			resource VARCHAR(255) NOT NULL,
			resource_id UUID,
			details JSONB,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		`INSERT INTO roles (name, description) VALUES 
			('admin', 'Full system administrator') 
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO roles (name, description) VALUES 
			('user', 'Regular user with limited access') 
			ON CONFLICT (name) DO NOTHING;`,

		`INSERT INTO permissions (name, resource, action) VALUES 
			('users.read', 'users', 'read'),
			('users.write', 'users', 'write'),
			('roles.read', 'roles', 'read'),
			('roles.write', 'roles', 'write'),
			('secrets.read', 'secrets', 'read'),
			('secrets.write', 'secrets', 'write'),
			('audit.read', 'audit', 'read')
			ON CONFLICT (name) DO NOTHING;`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %v", err)
		}
	}

	return nil
}

// docker exec -it miniidam-pamplatform-fullstack-postgres-1 psql -U postgres -d idam_pam
