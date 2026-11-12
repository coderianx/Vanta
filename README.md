# Vanta

Vanta is a REST API project built with Go, PostgreSQL, and JWT, **with the vision of becoming a massive community platform**. It currently includes authentication components for user registration and login.

## Vision

Vanta aims to create a large-scale community platform where people can come together around their interests, build their own communities, share content, and interact with one another.

The goal is to grow into a platform capable of supporting a high volume of users and content through features such as user profiles, communities, posts, comments, interactions, and moderation systems. The current authentication infrastructure is one of the first steps toward this vision.

> **Status:** The project is under development. Registration, login, and refresh token rotation handlers have been implemented; however, database initialization and HTTP server setup in `cmd/vanta/main.go` have not yet been completed.

## Current Features

- Secure password hashing and verification with bcrypt
- User registration and queries with PostgreSQL
- Access token generation with JWT, valid for 24 hours
- Refresh token generation, valid for 7 days, with only its SHA-256 hash stored in the database
- Atomic refresh token rotation that invalidates the old token and issues a new token pair
- Bearer JWT authentication middleware with request-scoped claims access
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

### Refresh Token Rotation

```http
POST /api/v1/auth/refresh
Content-Type: application/json
```

```json
{
  "refresh_token": "current-refresh-token"
}
```

A valid refresh token is deleted and replaced with a new one in a single database transaction. The response contains a new access token and refresh token:

```json
{
  "token": "new-signed-jwt-token",
  "refresh_token": "new-random-refresh-token"
}
```

The old refresh token cannot be used again. An unknown, expired, or previously rotated token returns `401 Unauthorized` with an `invalid_refresh_token` response.

### Current User

```http
GET /api/v1/me
Authorization: Bearer signed-jwt-token
```

The protected endpoint reads the user ID from the access token claims and returns the user's current account data together with their follower and following counts:

```json
{
  "id": 1,
  "name": "Ada Lovelace",
  "username": "ada",
  "follower_count": 120,
  "following_count": 80
}
```

The password hash is never included in the response. A missing or invalid access token returns `401 Unauthorized`; if the user no longer exists, the handler returns `404 Not Found` with a `user_not_found` response.

### Follow a User

```http
POST /api/v1/users/{username}/follow
Authorization: Bearer signed-jwt-token
```

A new follow relationship returns `201 Created`. If the relationship already exists, the operation remains idempotent and returns `200 OK`:

```json
{
  "following": true,
  "username": "ada"
}
```

Attempting to follow yourself returns `400 Bad Request` with a `cannot_follow_self` response. An unknown target username returns `404 Not Found`.

### Unfollow a User

```http
DELETE /api/v1/users/{username}/follow
Authorization: Bearer signed-jwt-token
```

The operation is idempotent and returns `200 OK` whether or not the follow relationship existed:

```json
{
  "following": false,
  "username": "ada"
}
```

### Follow Status

```http
GET /api/v1/users/{username}/follow-status
Authorization: Bearer signed-jwt-token
```

The response shows both follow directions and whether the relationship is mutual:

```json
{
  "following": true,
  "followed_by": false,
  "mutual": false,
  "username": "ada"
}
```

### Public User Profile

```http
GET /api/v1/users/{username}
```

This endpoint does not require authentication. It returns public account data and follow counts without exposing the password hash:

```json
{
  "id": 1,
  "name": "Ada Lovelace",
  "username": "ada",
  "follower_count": 120,
  "following_count": 80
}
```

An unknown username returns `404 Not Found` with a `user_not_found` response.

## Protecting Routes

Use `AuthMiddleware` with Chi's `r.With(...)` method to protect a route:

```go
r.With(middlewares.AuthMiddleware).Get("/me", auth.GetMe)
```

Clients must send the access token in the Authorization header:

```http
Authorization: Bearer signed-jwt-token
```

The middleware accepts HS256 tokens issued by `vanta` and requires a valid expiration claim. Inside a protected handler, retrieve the claims with `GetClaims`:

```go
claims, ok := middlewares.GetClaims(r)

if !ok {
    http.Error(w, "unauthorized", http.StatusUnauthorized)
    return
}

userID := claims.UserID
username := claims.Username
```

Missing, invalid, or expired access tokens are rejected with `401 Unauthorized` and an `invalid_token` response.

## Running

Once the application startup flow is complete, the project can be run with:

```bash
go run ./cmd/vanta
```

Currently, `main.go` does not call `store.ConnectDatabase()` or `http.ListenAndServe()`, so this command does not start an HTTP server.

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
│   ├── handlers/auth/         # Authentication HTTP handlers
│   ├── handlers/follow/       # Follow relationship HTTP handlers
│   ├── handlers/users/        # Public user profile HTTP handlers
│   ├── middlewares/           # JWT authentication middleware
│   ├── models/                # Request, user, and JWT models
│   ├── repository/            # User database operations
│   └── store/                 # PostgreSQL connection and table setup
├── go.mod
├── go.sum
└── README.md
```

## License

This project is licensed under the [MIT License](LICENSE).
