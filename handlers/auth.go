package handlers

import (
	"net/http"
	"time"

	"github.com/devreserve/server/config"
	"github.com/devreserve/server/db"
	"github.com/devreserve/server/models"
	"github.com/devreserve/server/utils"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo *db.UserRepository
	config   config.Config
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *db.UserRepository, config config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		config:   config,
	}
}

// Register handles user registration requests
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse the request body
	var req models.RegisterRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the username and password
	if req.Username == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Username is required")
		return
	}
	if len(req.Password) < 8 {
		utils.RespondWithError(w, http.StatusBadRequest, "Password must be at least 8 characters")
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
		Role:        models.RoleUser, // By default, new users are regular users
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}

	if err := h.userRepo.CreateUser(user); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate a token for the new user
	token, err := utils.GenerateToken(user, h.config)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Respond with the token
	utils.RespondWithSuccess(w, map[string]interface{}{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// Login handles user login requests
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse the request body
	var req models.LoginRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the username and password
	if req.Username == "" || req.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get the user
	user, err := h.userRepo.GetUser(req.Username)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if user == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Check the password
	if !utils.CheckPassword(req.Password, user.Password) {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate a token
	token, err := utils.GenerateToken(*user, h.config)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Respond with the token
	utils.RespondWithSuccess(w, map[string]interface{}{
		"token": token,
		"user":  user.ToResponse(),
	})
}
