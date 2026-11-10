# Vanta

Vanta is a REST API project built with Go, PostgreSQL, and JWT, **with the vision of becoming a massive community platform**. It currently includes authentication components for user registration and login.

## Vision

Vanta aims to create a large-scale community platform where people can come together around their interests, build their own communities, share content, and interact with one another.

The goal is to grow into a platform capable of supporting a high volume of users and content through features such as user profiles, communities, posts, comments, interactions, and moderation systems. The current authentication infrastructure is one of the first steps toward this vision.

> **Status:** The project is under development. Registration and login handlers have been implemented; however, database initialization, route registrations, and HTTP server setup in `cmd/vanta/main.go` have not yet been completed.

## Current Features

- Secure password hashing and verification with bcrypt
- User registration and queries with PostgreSQL
- Access token generation with JWT, valid for 24 hours
- Refresh token generation, valid for 7 days, with only its SHA-256 hash stored in the database
- Route organization under `/api/v1` with Chi router
- Separation of handler, repository, model, and database layers
- Parameterized SQL queries and username uniqueness checks

## Technologies

- Go 1.27.1
- PostgreSQL
- Chi
- pgx
- golang-jwt
- bcrypt
- godotenv

## Requirements

- Go 1.27.1 or a compatible version
- A running PostgreSQL database

## Installation

Clone the repository and download the dependencies:

```bash
git clone https://github.com/coderianx/Vanta.git
cd Vanta
go mod download
```

Create a `.env` file in the project root:

```env
DB_URL=postgres://postgres:postgres@localhost:5432/vanta?sslmode=disable
JWT_SECRET=a-strong-and-random-secret
```

Replace `JWT_SECRET` with a long, random, and unpredictable value in production environments. The `.env` file is excluded by `.gitignore` and must not be added to version control.

## Database

The application uses the following `users` and `refresh_tokens` tables:

```sql
CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

The tables can be created manually using the SQL above or by calling `store.CreateTables()` during application startup. Only the SHA-256 hash of the refresh token is stored; the raw token is sent to the client only in the login response.

## API

### User Registration

```http
POST /api/v1/auth/register
Content-Type: application/json
```

```json
{
  "name": "Ada Lovelace",
  "username": "ada",
  "password": "a-strong-password"
}
```

The password must be between 8 and 72 bytes long. Successful registration returns `201 Created`:

```json
{
  "message": "user registered successfully"
}
```

If the username is already registered, `409 Conflict` is returned.

### User Login

```http
POST /api/v1/auth/login
Content-Type: application/json
```

```json
{
  "username": "ada",
  "password": "a-strong-password"
}
```

A successful login returns `200 OK` with a JWT and a refresh token valid for 7 days:

```json
{
  "token": "signed-jwt-token",
  "refresh_token": "random-refresh-token"
}
```

An invalid username or password returns `401 Unauthorized` with an `invalid_credentials` response.

## Running

Once the application startup flow is complete, the project can be run with:

```bash
go run ./cmd/vanta
```

Currently, `main.go` does not call `store.ConnectDatabase()` or `http.ListenAndServe()`, so this command does not start an HTTP server. Additionally, the `/api/v1/auth/login` route must be registered with `auth.LoginHandler` for the login endpoint to be available.

## Checks

```bash
gofmt -w .
go test ./...
go vet ./...
go build ./...
```

## Project Structure

```text
.
├── cmd/vanta/                 # Application entry point
├── internal/
│   ├── handlers/auth/         # Registration and login HTTP handlers
│   ├── models/                # Request, user, and JWT models
│   ├── repository/            # User database operations
│   └── store/                 # PostgreSQL connection and table setup
├── go.mod
├── go.sum
└── README.md
```

## License

This project is licensed under the [MIT License](LICENSE).
