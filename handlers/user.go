package handlers

import (
	"net/http"
	"time"

	"github.com/devreserve/server/db"
	"github.com/devreserve/server/middleware"
	"github.com/devreserve/server/models"
	"github.com/devreserve/server/utils"
	"github.com/gorilla/mux"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo *db.UserRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepo *db.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// ListUsers handles requests to list all users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get all users
	users, err := h.userRepo.ListUsers()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	// Respond with the users
	utils.RespondWithSuccess(w, users)
}

// CreateUser handles requests to create a new user (admin only)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the admin user from the request context
	userValue := r.Context().Value(middleware.UserContextKey)
	if userValue == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	admin, ok := userValue.(models.User)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid user context")
		return
	}

	// Verify that the user is an admin
	if admin.Role != models.RoleAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Parse the request body
	var req struct {
		Username string         `json:"username"`
		Password string         `json:"password"`
		Role     models.UserRole `json:"role"`
	}
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the request
	if req.Username == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Username is required")
		return
	}
	if len(req.Password) < 8 {
		utils.RespondWithError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}
	if req.Role == "" {
		req.Role = models.RoleUser
	}
	if req.Role != models.RoleUser && req.Role != models.RoleAdmin {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid role")
		return
	}

	// Check if the username already exists
	existingUser, err := h.userRepo.GetUser(req.Username)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check username")
		return
	}
	if existingUser != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Username already exists")
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create the user
	user := models.User{
		Username:    req.Username,
		Password:    hashedPassword,
		Role:        req.Role,
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}

	if err := h.userRepo.CreateUser(user); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Respond with the created user
	utils.RespondWithSuccess(w, user.ToResponse())
}

// GetUser handles requests to get a user by username
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the username from the URL parameters
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Username is required")
		return
	}

	// Get the user
	user, err := h.userRepo.GetUser(username)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if user == nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Respond with the user
	utils.RespondWithSuccess(w, user.ToResponse())
}
