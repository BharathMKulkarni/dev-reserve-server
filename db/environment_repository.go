package db

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/devreserve/server/models"
	"github.com/google/uuid"
)

// EnvironmentRepository handles operations on the Environments table
type EnvironmentRepository struct {
	db *DynamoDBClient
}

// NewEnvironmentRepository creates a new EnvironmentRepository
func NewEnvironmentRepository(db *DynamoDBClient) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}

// CreateEnvironment creates a new environment in the database
func (r *EnvironmentRepository) CreateEnvironment(env models.Environment, username string) (*models.Environment, error) {
	// Generate a new ID for the environment
	env.ID = uuid.New().String()
	env.Status = models.StatusFree
	env.CreatedBy = username

	// Set the timestamps
	now := time.Now()
	env.CreatedAt = now
	env.LastUpdated = now

	// Convert the environment to a DynamoDB item
	item, err := dynamodbattribute.MarshalMap(env)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal environment: %w", err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String(EnvironmentsTableName),
		Item:      item,
	}

	// Put the item in DynamoDB
	_, err = r.db.Client.PutItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	return &env, nil
}

// GetEnvironment gets an environment by ID
func (r *EnvironmentRepository) GetEnvironment(id string) (*models.Environment, error) {
	// Create the input for the GetItem operation
	input := &dynamodb.GetItemInput{
		TableName: aws.String(EnvironmentsTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	// Get the item from DynamoDB
	result, err := r.db.Client.GetItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}

	// Check if the item exists
	if result.Item == nil {
		return nil, nil
	}

	// Unmarshal the item into an Environment struct
	var env models.Environment
	err = dynamodbattribute.UnmarshalMap(result.Item, &env)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal environment: %w", err)
	}

	return &env, nil
}

// ListEnvironments gets all environments
func (r *EnvironmentRepository) ListEnvironments() ([]models.Environment, error) {
	// Create the input for the Scan operation
	input := &dynamodb.ScanInput{
		TableName: aws.String(EnvironmentsTableName),
	}

	// Scan the table
	result, err := r.db.Client.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	// Unmarshal the items into Environment structs
	var environments []models.Environment
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &environments)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal environments: %w", err)
	}

	return environments, nil
}

// UpdateEnvironment updates an existing environment
func (r *EnvironmentRepository) UpdateEnvironment(env models.Environment) error {
	// Set the last updated timestamp
	env.LastUpdated = time.Now()

	// Convert the environment to a DynamoDB item
	item, err := dynamodbattribute.MarshalMap(env)
	if err != nil {
		return fmt.Errorf("failed to marshal environment: %w", err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String(EnvironmentsTableName),
		Item:      item,
		// Ensure the environment ID exists
		ConditionExpression: aws.String("attribute_exists(id)"),
	}

	// Put the item in DynamoDB
	_, err = r.db.Client.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to update environment: %w", err)
	}

	return nil
}

// UpdateEnvironmentStatus updates the status of an environment
func (r *EnvironmentRepository) UpdateEnvironmentStatus(id string, status models.EnvironmentStatus) error {
	// Create the input for the UpdateItem operation
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(EnvironmentsTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String("SET #status = :status, #lastUpdated = :lastUpdated"),
		ExpressionAttributeNames: map[string]*string{
			"#status":      aws.String("status"),
			"#lastUpdated": aws.String("lastUpdated"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":status": {
				S: aws.String(string(status)),
			},
			":lastUpdated": {
				S: aws.String(time.Now().Format(time.RFC3339)),
			},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
	}

	// Update the item in DynamoDB
	_, err := r.db.Client.UpdateItem(input)
	if err != nil {
		return fmt.Errorf("failed to update environment status: %w", err)
	}

	return nil
}

// DeleteEnvironment deletes an environment by ID
func (r *EnvironmentRepository) DeleteEnvironment(id string) error {
	// Create the input for the DeleteItem operation
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(EnvironmentsTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	// Delete the item from DynamoDB
	_, err := r.db.Client.DeleteItem(input)
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	return nil
}
