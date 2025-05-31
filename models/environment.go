package models

import (
	"time"
)

// EnvironmentStatus represents the current status of an environment
type EnvironmentStatus string

const (
	// StatusFree indicates that the environment is available for reservation
	StatusFree EnvironmentStatus = "FREE"
	// StatusReserved indicates that the environment is currently reserved
	StatusReserved EnvironmentStatus = "RESERVED"
)

// Environment represents a testing environment that can be reserved by users
type Environment struct {
	ID          string            `json:"id" dynamodbav:"id"`
	Name        string            `json:"name" dynamodbav:"name"`
	Description string            `json:"description,omitempty" dynamodbav:"description"`
	Status      EnvironmentStatus `json:"status" dynamodbav:"status"`
	CreatedBy   string            `json:"createdBy" dynamodbav:"createdBy"`
	CreatedAt   time.Time         `json:"createdAt" dynamodbav:"createdAt"`
	LastUpdated time.Time         `json:"lastUpdated" dynamodbav:"lastUpdated"`
}

// EnvironmentCreateRequest represents the data needed to create a new environment
type EnvironmentCreateRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
}

// Reservation represents a reservation of an environment by a user
type Reservation struct {
	ID            string    `json:"id" dynamodbav:"id"`
	EnvironmentID string    `json:"environmentId" dynamodbav:"environmentId"`
	Username      string    `json:"username" dynamodbav:"username"`
	StartTime     time.Time `json:"startTime" dynamodbav:"startTime"`
	EndTime       time.Time `json:"endTime" dynamodbav:"endTime"`
	Feature       string    `json:"feature" dynamodbav:"feature"`
	GitBranch     string    `json:"gitBranch,omitempty" dynamodbav:"gitBranch,omitempty"`
	JiraURL       string    `json:"jiraUrl,omitempty" dynamodbav:"jiraUrl,omitempty"`
	CreatedAt     time.Time `json:"createdAt" dynamodbav:"createdAt"`
	LastUpdated   time.Time `json:"lastUpdated" dynamodbav:"lastUpdated"`
}

// ReservationCreateRequest represents the data needed to create a new reservation
type ReservationCreateRequest struct {
	EnvironmentID string `json:"environmentId" validate:"required"`
	DurationMins  int    `json:"durationMins" validate:"required,min=10,max=4320"` // Min 10 mins, Max 3 days (4320 mins)
	Feature       string `json:"feature" validate:"required"`
	GitBranch     string `json:"gitBranch,omitempty"`
	JiraURL       string `json:"jiraUrl,omitempty"`
}

// EnvironmentWithReservation represents an environment with its current reservation (if any)
type EnvironmentWithReservation struct {
	Environment
	CurrentReservation *Reservation `json:"currentReservation,omitempty"`
}
