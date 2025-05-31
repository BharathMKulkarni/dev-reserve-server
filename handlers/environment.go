package handlers

import (
	"net/http"

	"github.com/devreserve/server/db"
	"github.com/devreserve/server/middleware"
	"github.com/devreserve/server/models"
	"github.com/devreserve/server/utils"
	"github.com/gorilla/mux"
)

// EnvironmentHandler handles environment-related requests
type EnvironmentHandler struct {
	envRepo        *db.EnvironmentRepository
	reservationRepo *db.ReservationRepository
}

// NewEnvironmentHandler creates a new EnvironmentHandler
func NewEnvironmentHandler(envRepo *db.EnvironmentRepository, reservationRepo *db.ReservationRepository) *EnvironmentHandler {
	return &EnvironmentHandler{
		envRepo:        envRepo,
		reservationRepo: reservationRepo,
	}
}

// ListEnvironments handles requests to list all environments
func (h *EnvironmentHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get all environments
	environments, err := h.envRepo.ListEnvironments()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list environments")
		return
	}

	// Get all active reservations
	activeReservations, err := h.reservationRepo.ListActiveReservations()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list reservations")
		return
	}

	// Create a map of environment ID to reservation
	reservationMap := make(map[string]*models.Reservation)
	for i := range activeReservations {
		reservationMap[activeReservations[i].EnvironmentID] = &activeReservations[i]
	}

	// Create environment with reservation response objects
	result := make([]models.EnvironmentWithReservation, len(environments))
	for i, env := range environments {
		result[i].Environment = env
		result[i].CurrentReservation = reservationMap[env.ID]
	}

	// Respond with the environments
	utils.RespondWithSuccess(w, result)
}

// CreateEnvironment handles requests to create a new environment
func (h *EnvironmentHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the user from the request context
	userValue := r.Context().Value(middleware.UserContextKey)
	if userValue == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	user, ok := userValue.(models.User)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid user context")
		return
	}

	// Parse the request body
	var req models.EnvironmentCreateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the environment name
	if req.Name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Environment name is required")
		return
	}

	// Create the environment
	env := models.Environment{
		Name:        req.Name,
		Description: req.Description,
		Status:      models.StatusFree,
	}

	createdEnv, err := h.envRepo.CreateEnvironment(env, user.Username)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create environment")
		return
	}

	// Respond with the created environment
	utils.RespondWithSuccess(w, createdEnv)
}

// GetEnvironment handles requests to get an environment by ID
func (h *EnvironmentHandler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the environment ID from the URL parameters
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Environment ID is required")
		return
	}

	// Get the environment
	env, err := h.envRepo.GetEnvironment(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get environment")
		return
	}
	if env == nil {
		utils.RespondWithError(w, http.StatusNotFound, "Environment not found")
		return
	}

	// Get the active reservation for the environment
	reservation, err := h.reservationRepo.GetActiveReservationByEnvironmentID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get reservation")
		return
	}

	// Create the response
	result := models.EnvironmentWithReservation{
		Environment:        *env,
		CurrentReservation: reservation,
	}

	// Respond with the environment
	utils.RespondWithSuccess(w, result)
}
