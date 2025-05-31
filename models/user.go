package models

import (
	"time"
)

// UserRole defines the role of a user in the system
type UserRole string

const (
	// RoleAdmin represents an admin user who can manage users and environments
	RoleAdmin UserRole = "ADMIN"
	// RoleUser represents a normal user who can reserve environments
	RoleUser UserRole = "USER"
)

// User represents a developer in the system who can reserve environments
type User struct {
	Username    string    `json:"username" dynamodbav:"username"`
	Password    string    `json:"-" dynamodbav:"password"` // Password is not returned in JSON responses
	Role        UserRole  `json:"role" dynamodbav:"role"`
	CreatedAt   time.Time `json:"createdAt" dynamodbav:"createdAt"`
	LastUpdated time.Time `json:"lastUpdated" dynamodbav:"lastUpdated"`
}

// UserResponse is used for returning user data in API responses (without the password)
type UserResponse struct {
	Username    string    `json:"username"`
	Role        UserRole  `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// ToResponse converts a User to a UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		Username:    u.Username,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt,
		LastUpdated: u.LastUpdated,
	}
}

// LoginRequest represents the data needed for user login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents the data needed for user registration
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
