package db

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/devreserve/server/models"
	"github.com/google/uuid"
)

// ReservationRepository handles operations on the Reservations table
type ReservationRepository struct {
	db *DynamoDBClient
	envRepo *EnvironmentRepository
}

// NewReservationRepository creates a new ReservationRepository
func NewReservationRepository(db *DynamoDBClient, envRepo *EnvironmentRepository) *ReservationRepository {
	return &ReservationRepository{
		db: db,
		envRepo: envRepo,
	}
}

// CreateReservation creates a new reservation in the database
func (r *ReservationRepository) CreateReservation(reservation models.Reservation) (*models.Reservation, error) {
	// Get the environment to check if it's available
	env, err := r.envRepo.GetEnvironment(reservation.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return nil, fmt.Errorf("environment not found")
	}
	if env.Status != models.StatusFree {
		return nil, fmt.Errorf("environment is already reserved")
	}

	// Generate a new ID for the reservation
	reservation.ID = uuid.New().String()

	// Set the timestamps
	now := time.Now()
	reservation.CreatedAt = now
	reservation.LastUpdated = now

	// Convert the reservation to a DynamoDB item
	item, err := dynamodbattribute.MarshalMap(reservation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reservation: %w", err)
	}

	// Create a transaction to create the reservation and update the environment status
	// First, prepare the transaction item for creating the reservation
	putReservation := &dynamodb.TransactWriteItem{
		Put: &dynamodb.Put{
			TableName: aws.String(ReservationsTableName),
			Item:      item,
		},
	}

	// Second, prepare the transaction item for updating the environment status
	updateEnv := &dynamodb.TransactWriteItem{
		Update: &dynamodb.Update{
			TableName: aws.String(EnvironmentsTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(reservation.EnvironmentID),
				},
			},
			UpdateExpression: aws.String("SET #status = :status, #lastUpdated = :lastUpdated"),
			ExpressionAttributeNames: map[string]*string{
				"#status":      aws.String("status"),
				"#lastUpdated": aws.String("lastUpdated"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":status": {
					S: aws.String(string(models.StatusReserved)),
				},
				":lastUpdated": {
					S: aws.String(now.Format(time.RFC3339)),
				},
				":expectedStatus": {
					S: aws.String(string(models.StatusFree)),
				},
			},
			ConditionExpression: aws.String("#status = :expectedStatus"),
		},
	}

	// Execute the transaction
	_, err = r.db.Client.TransactWriteItems(&dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			putReservation,
			updateEnv,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	return &reservation, nil
}

// GetReservation gets a reservation by ID
func (r *ReservationRepository) GetReservation(id string) (*models.Reservation, error) {
	// Create the input for the GetItem operation
	input := &dynamodb.GetItemInput{
		TableName: aws.String(ReservationsTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	// Get the item from DynamoDB
	result, err := r.db.Client.GetItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	// Check if the item exists
	if result.Item == nil {
		return nil, nil
	}

	// Unmarshal the item into a Reservation struct
	var reservation models.Reservation
	err = dynamodbattribute.UnmarshalMap(result.Item, &reservation)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reservation: %w", err)
	}

	return &reservation, nil
}

// GetActiveReservationByEnvironmentID gets the active reservation for an environment
func (r *ReservationRepository) GetActiveReservationByEnvironmentID(environmentID string) (*models.Reservation, error) {
	// Create a filter expression for active reservations
	now := time.Now()
	filt := expression.And(
		expression.Name("environmentId").Equal(expression.Value(environmentID)),
		expression.Name("endTime").GreaterThan(expression.Value(now.Format(time.RFC3339))),
	)
	
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	// Create the input for the Scan operation
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(ReservationsTableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	// Scan the table
	result, err := r.db.Client.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan for active reservations: %w", err)
	}

	// Check if any active reservations were found
	if len(result.Items) == 0 {
		return nil, nil
	}

	// Unmarshal the item into a Reservation struct
	var reservations []models.Reservation
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &reservations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reservations: %w", err)
	}

	// Return the first (and should be only) active reservation
	return &reservations[0], nil
}

// ListActiveReservations gets all currently active reservations
func (r *ReservationRepository) ListActiveReservations() ([]models.Reservation, error) {
	// Create a filter expression for active reservations
	now := time.Now()
	filt := expression.Name("endTime").GreaterThan(expression.Value(now.Format(time.RFC3339)))
	
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	// Create the input for the Scan operation
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(ReservationsTableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	// Scan the table
	result, err := r.db.Client.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan for active reservations: %w", err)
	}

	// Check if any active reservations were found
	if len(result.Items) == 0 {
		return []models.Reservation{}, nil
	}

	// Unmarshal the items into Reservation structs
	var reservations []models.Reservation
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &reservations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reservations: %w", err)
	}

	return reservations, nil
}

// ReleaseReservation releases a reservation before its end time
func (r *ReservationRepository) ReleaseReservation(id string, username string) error {
	// Get the reservation to check if it exists and belongs to the user
	reservation, err := r.GetReservation(id)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}
	if reservation == nil {
		return fmt.Errorf("reservation not found")
	}
	if reservation.Username != username {
		return fmt.Errorf("you can only release your own reservations")
	}

	// Create a transaction to update the reservation's end time and the environment status
	// First, prepare the transaction item for updating the reservation
	updateReservation := &dynamodb.TransactWriteItem{
		Update: &dynamodb.Update{
			TableName: aws.String(ReservationsTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(id),
				},
			},
			UpdateExpression: aws.String("SET #endTime = :endTime, #lastUpdated = :lastUpdated"),
			ExpressionAttributeNames: map[string]*string{
				"#endTime":     aws.String("endTime"),
				"#lastUpdated": aws.String("lastUpdated"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":endTime": {
					S: aws.String(time.Now().Format(time.RFC3339)),
				},
				":lastUpdated": {
					S: aws.String(time.Now().Format(time.RFC3339)),
				},
			},
		},
	}

	// Second, prepare the transaction item for updating the environment status
	updateEnv := &dynamodb.TransactWriteItem{
		Update: &dynamodb.Update{
			TableName: aws.String(EnvironmentsTableName),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(reservation.EnvironmentID),
				},
			},
			UpdateExpression: aws.String("SET #status = :status, #lastUpdated = :lastUpdated"),
			ExpressionAttributeNames: map[string]*string{
				"#status":      aws.String("status"),
				"#lastUpdated": aws.String("lastUpdated"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":status": {
					S: aws.String(string(models.StatusFree)),
				},
				":lastUpdated": {
					S: aws.String(time.Now().Format(time.RFC3339)),
				},
			},
		},
	}

	// Execute the transaction
	_, err = r.db.Client.TransactWriteItems(&dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			updateReservation,
			updateEnv,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to release reservation: %w", err)
	}

	return nil
}

// CheckExpiredReservations checks for expired reservations and updates their environments to be free
func (r *ReservationRepository) CheckExpiredReservations() error {
	// Get all active reservations
	activeReservations, err := r.ListActiveReservations()
	if err != nil {
		return fmt.Errorf("failed to list active reservations: %w", err)
	}

	now := time.Now()
	
	// Check each reservation to see if it's expired
	for _, reservation := range activeReservations {
		if now.After(reservation.EndTime) {
			// The reservation has expired, update the environment status
			err := r.envRepo.UpdateEnvironmentStatus(reservation.EnvironmentID, models.StatusFree)
			if err != nil {
				return fmt.Errorf("failed to update environment status: %w", err)
			}
		}
	}

	return nil
}
