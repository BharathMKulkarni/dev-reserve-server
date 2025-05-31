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

// ReservationHandler handles reservation-related requests
type ReservationHandler struct {
	reservationRepo *db.ReservationRepository
	envRepo         *db.EnvironmentRepository
}

// NewReservationHandler creates a new ReservationHandler
func NewReservationHandler(reservationRepo *db.ReservationRepository, envRepo *db.EnvironmentRepository) *ReservationHandler {
	return &ReservationHandler{
		reservationRepo: reservationRepo,
		envRepo:         envRepo,
	}
}

// CreateReservation handles requests to reserve an environment
func (h *ReservationHandler) CreateReservation(w http.ResponseWriter, r *http.Request) {
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
	var req models.ReservationCreateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the request
	if req.EnvironmentID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Environment ID is required")
		return
	}
	if req.DurationMins < 10 {
		utils.RespondWithError(w, http.StatusBadRequest, "Duration must be at least 10 minutes")
		return
	}
	if req.DurationMins > 4320 { // 3 days = 4320 minutes
		utils.RespondWithError(w, http.StatusBadRequest, "Duration cannot exceed 3 days (4320 minutes)")
		return
	}
	if req.Feature == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Feature description is required")
		return
	}

	// Get the environment to check if it's available
	env, err := h.envRepo.GetEnvironment(req.EnvironmentID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get environment")
		return
	}
	if env == nil {
		utils.RespondWithError(w, http.StatusNotFound, "Environment not found")
		return
	}
	if env.Status != models.StatusFree {
		utils.RespondWithError(w, http.StatusBadRequest, "Environment is already reserved")
		return
	}

	// Create the reservation
	now := time.Now()
	endTime := now.Add(time.Duration(req.DurationMins) * time.Minute)
	
	reservation := models.Reservation{
		EnvironmentID: req.EnvironmentID,
		Username:      user.Username,
		StartTime:     now,
		EndTime:       endTime,
		Feature:       req.Feature,
		GitBranch:     req.GitBranch,
		JiraURL:       req.JiraURL,
	}

	createdReservation, err := h.reservationRepo.CreateReservation(reservation)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create reservation: "+err.Error())
		return
	}

	// Respond with the created reservation
	utils.RespondWithSuccess(w, createdReservation)
}

// ReleaseReservation handles requests to release a reserved environment
func (h *ReservationHandler) ReleaseReservation(w http.ResponseWriter, r *http.Request) {
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

	// Get the reservation ID from the URL parameters
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Reservation ID is required")
		return
	}

	// Release the reservation
	err := h.reservationRepo.ReleaseReservation(id, user.Username)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to release reservation: "+err.Error())
		return
	}

	// Respond with success
	utils.RespondWithSuccess(w, map[string]interface{}{
		"message": "Reservation released successfully",
	})
}

// GetActiveReservations handles requests to get all active reservations
func (h *ReservationHandler) GetActiveReservations(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get all active reservations
	reservations, err := h.reservationRepo.ListActiveReservations()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list reservations")
		return
	}

	// Respond with the reservations
	utils.RespondWithSuccess(w, reservations)
}
