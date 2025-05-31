package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/devreserve/server/config"
	"github.com/devreserve/server/db"
	"github.com/devreserve/server/handlers"
	"github.com/devreserve/server/middleware"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load the .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables")
	} else {
		log.Println("Loaded environment variables from .env file")
	}

	// Load the application configuration
	cfg := config.LoadConfig()

	// Create the DynamoDB client
	dbClient, err := db.NewDynamoDBClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	// Ensure the required tables exist
	if err := dbClient.CreateTablesIfNotExist(); err != nil {
		log.Fatalf("Failed to create DynamoDB tables: %v", err)
	}

	// Create the repositories
	userRepo := db.NewUserRepository(dbClient)
	envRepo := db.NewEnvironmentRepository(dbClient)
	reservationRepo := db.NewReservationRepository(dbClient, envRepo)

	// Create the handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	userHandler := handlers.NewUserHandler(userRepo)
	envHandler := handlers.NewEnvironmentHandler(envRepo, reservationRepo)
	reservationHandler := handlers.NewReservationHandler(reservationRepo, envRepo)

	// Create the router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// Protected routes
	authRouter := router.PathPrefix("/api").Subrouter()
	authRouter.Use(middleware.AuthMiddleware(cfg))

	// User routes
	authRouter.HandleFunc("/users", userHandler.ListUsers).Methods("GET")
	authRouter.HandleFunc("/users/{username}", userHandler.GetUser).Methods("GET")

	// Admin-only routes
	adminRouter := authRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminMiddleware)
	adminRouter.HandleFunc("/users", userHandler.CreateUser).Methods("POST")

	// Environment routes
	authRouter.HandleFunc("/environments", envHandler.ListEnvironments).Methods("GET")
	authRouter.HandleFunc("/environments/{id}", envHandler.GetEnvironment).Methods("GET")
	adminRouter.HandleFunc("/environments", envHandler.CreateEnvironment).Methods("POST")

	// Reservation routes
	authRouter.HandleFunc("/reservations", reservationHandler.CreateReservation).Methods("POST")
	authRouter.HandleFunc("/reservations", reservationHandler.GetActiveReservations).Methods("GET")
	authRouter.HandleFunc("/reservations/{id}/release", reservationHandler.ReleaseReservation).Methods("POST")

	// Set up CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // You should restrict this in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Start a background goroutine to check for expired reservations
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := reservationRepo.CheckExpiredReservations(); err != nil {
				log.Printf("Error checking expired reservations: %v", err)
			}
		}
	}()

	// Create the server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      corsMiddleware.Handler(router),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for an interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	log.Println("Server shutting down...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}
