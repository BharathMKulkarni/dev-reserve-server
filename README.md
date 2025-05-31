# Dev Reserve Server

A backend API for managing development environment reservations, built with Go and DynamoDB.

## Overview

The Dev Reserve Server provides a REST API that allows development teams to manage and reserve environments for testing. It supports:

- User authentication (login, registration)
- Environment management (create, list, get)
- Environment reservation (reserve, release)
- Automatic release of environments when reservation time expires

## Architecture

The application follows a clean architecture approach with the following components:

- **Models**: Data structures and business logic
- **Handlers**: HTTP request handlers
- **Middleware**: Authentication and authorization
- **DB**: Database access layer
- **Utils**: Utility functions (password hashing, JWT, etc.)
- **Config**: Application configuration

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get a JWT token

### Users

- `GET /api/users` - List all users (authenticated)
- `GET /api/users/{username}` - Get a user by username (authenticated)
- `POST /api/admin/users` - Create a new user (admin only)

### Environments

- `GET /api/environments` - List all environments (authenticated)
- `GET /api/environments/{id}` - Get an environment by ID (authenticated)
- `POST /api/admin/environments` - Create a new environment (admin only)

### Reservations

- `GET /api/reservations` - List all active reservations (authenticated)
- `POST /api/reservations` - Create a new reservation (authenticated)
- `POST /api/reservations/{id}/release` - Release a reservation (authenticated, owner only)

## Setup and Installation

### Prerequisites

- Go 1.19 or higher
- DynamoDB (local or AWS)
- AWS CLI (for local DynamoDB setup)

### Environment Variables

The application can be configured using the following environment variables:

- `PORT` - Server port (default: 8080)
- `AWS_REGION` - AWS region (default: us-east-1)
- `DYNAMODB_ENDPOINT` - DynamoDB endpoint (leave empty for AWS, set to `http://localhost:8000` for local)
- `JWT_SECRET` - Secret key for JWT token generation (default: dev-reserve-secret-key)

### Local Development

1. Set up a local DynamoDB instance:

```bash
docker run -p 8000:8000 amazon/dynamodb-local
```

2. Install dependencies:

```bash
go mod tidy
```

3. Run the application:

```bash
go run main.go
```

## Database Schema

### Users Table

- Primary Key: `username` (String)
- Attributes:
  - `password` (String)
  - `role` (String) - "ADMIN" or "USER"
  - `createdAt` (String - ISO8601)
  - `lastUpdated` (String - ISO8601)

### Environments Table

- Primary Key: `id` (String)
- Attributes:
  - `name` (String)
  - `description` (String)
  - `status` (String) - "FREE" or "RESERVED"
  - `createdBy` (String)
  - `createdAt` (String - ISO8601)
  - `lastUpdated` (String - ISO8601)

### Reservations Table

- Primary Key: `id` (String)
- GSI: `EnvironmentIndex` (environmentId)
- Attributes:
  - `environmentId` (String)
  - `username` (String)
  - `startTime` (String - ISO8601)
  - `endTime` (String - ISO8601)
  - `feature` (String)
  - `gitBranch` (String)
  - `jiraUrl` (String)
  - `createdAt` (String - ISO8601)
  - `lastUpdated` (String - ISO8601)

## API Authentication

The API uses JWT for authentication. After logging in, include the token in the Authorization header of subsequent requests:

```
Authorization: Bearer <token>
```

## Building and Deployment

Build the application:

```bash
go build -o dev-reserve-server
```

Run the compiled binary:

```bash
./dev-reserve-server
```
