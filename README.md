# ✨IDAM-PAM Platform

A comprehensive Identity and Access Management (IDAM) + Privileged Access Management (PAM) platform built with Go and React.

## 🚀 Features

### Backend (Go)

* **REST API** with Fiber framework
* **JWT Authentication** with Argon2 password hashing
* **Multi-Factor Authentication (MFA)** using TOTP (Google Authenticator)
* **Role-Based Access Control (RBAC)** with flexible permissions
* **Encrypted Credential Vault** with AES-GCM encryption
* **Comprehensive Audit Logging** to PostgreSQL
* **Modular Architecture** for maintainability

### Frontend (React)

* **Modern Admin Dashboard** with Tailwind CSS
* **User Management** interface
* **Secret Vault** management
* **Audit Log Viewer**
* **MFA Setup** integration
* **Responsive Design** for all devices

### Security Features

* Password hashing with Argon2
* JWT-based authentication
* TOTP-based two-factor authentication
* AES-GCM encryption for secrets
* Comprehensive audit logging
* Role-based permissions

## 🏗️ Architecture

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── auth/           # Authentication logic
│   ├── config/         # Configuration management
│   ├── database/       # Database operations
│   ├── encryption/     # Encryption services
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   └── server/         # Server setup
├── src/                # React frontend
│   ├── components/     # React components
│   ├── contexts/       # React contexts
│   ├── pages/          # Page components
│   └── services/       # API services
└── .github/workflows/  # CI/CD pipelines
```

## 🐳 Quick Start with Docker (uses your local Postgres)

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd idam-pam-platform
   ```

2. **Start backend and frontend with Docker Compose**

   * Ensure your local Postgres is running and accessible at `localhost:5432`
   * Compose uses `host.docker.internal` so containers can reach your local DB

   ```bash
   docker compose up -d --build backend frontend
   ```

3. **Access the application**

   * Frontend: [http://localhost:5173](http://localhost:5173)
   * Backend API: [http://localhost:5000](http://localhost:5000)
   * Health Check: [http://localhost:5000/health](http://localhost:5000/health)

Alternative without Compose (manual run):

```bash
# Build images
docker build -t miniidam-backend -f Dockerfile.backend .
docker build -t miniidam-frontend -f Dockerfile.frontend .

# Run backend (points to your local Postgres)
docker run -d --name miniidam-backend \
  -p 5000:5000 \
  -e PORT=5000 \
  -e DATABASE_URL="postgres://postgres:<YOUR_PASSWORD>@host.docker.internal:5432/idam_pam?sslmode=disable" \
  -e JWT_SECRET="dev-secret-change-later" \
  miniidam-backend

# Run frontend (static via nginx)
docker run -d --name miniidam-frontend -p 5173:80 miniidam-frontend
```

## 🔧 Local Development

### Prerequisites

* Go 1.21+
* Node.js 18+
* PostgreSQL 13+

### Backend Setup

1. **Install Go dependencies**

   ```bash
   go mod download
   ```

2. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Ensure PostgreSQL is running locally**

   * Create the database and extension once:

   ```bash
   psql -U postgres -h localhost -p 5432 -c "CREATE DATABASE idam_pam;" || true
   psql -U postgres -h localhost -p 5432 -d idam_pam -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
   ```

4. **Run the backend**

   ```bash
   go run cmd/server/main.go
   ```

### Frontend Setup

1. **Install dependencies**

   ```bash
   npm install
   ```

2. **Start the development server**

   ```bash
   npm run dev
   ```

## 🔐 Security Configuration

### Environment Variables

```env
# Database
DATABASE_URL=postgres://user:pass@host:port/dbname

# JWT
JWT_SECRET=your-super-secret-jwt-key

# Server
PORT=5000
```

## 🧪 Testing

### Run Backend Tests

```bash
go test -v ./...
```

### Run Frontend Tests

```bash
npm test
```

### Run Integration Tests

```bash
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## 📚 API Documentation

### Authentication Endpoints

* `POST /api/v1/auth/register` - Register new user
* `POST /api/v1/auth/login` - Login user
* `POST /api/v1/totp/enable` - Enable TOTP for user

### User Management

* `GET /api/v1/users` - List all users
* `GET /api/v1/users/:id` - Get user details
* `PUT /api/v1/users/:id` - Update user
* `POST /api/v1/users/:id/roles` - Assign role to user

### Secret Management

* `GET /api/v1/secrets` - List all secrets
* `POST /api/v1/secrets` - Create new secret
* `GET /api/v1/secrets/:id` - Get secret (decrypted)
* `DELETE /api/v1/secrets/:id` - Delete secret

### Audit Logs

* `GET /api/v1/audit` - Get audit logs

## 🚨 Production Considerations

### Security Checklist

* [ ] Change default JWT secret
* [ ] Use strong database passwords
* [ ] Enable SSL/TLS for database connections
* [ ] Configure proper CORS settings
* [ ] Set up rate limiting
* [ ] Use a secret manager for sensitive configuration
* [ ] Implement proper backup procedures
* [ ] Set up monitoring and alerting

### Performance

* [ ] Configure connection pooling for database
* [ ] Set up Redis for session caching
* [ ] Implement proper indexing in PostgreSQL
* [ ] Configure static asset caching (CDN or reverse proxy)

### Monitoring

* [ ] Centralized logging solution
* [ ] Configure application metrics
* [ ] Set up health check endpoints
* [ ] Implement distributed tracing

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:

* Create an issue on GitHub
* Check the documentation
* Review the audit logs for troubleshooting

---

**⚠️ Security Note**: This platform handles sensitive credentials and user data. Always follow security best practices and conduct regular security audits in production environments.

---
## Developed By Manikandan😎
