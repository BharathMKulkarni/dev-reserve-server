package db

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/devreserve/server/models"
)

// UserRepository handles operations on the Users table
type UserRepository struct {
	db *DynamoDBClient
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *DynamoDBClient) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user models.User) error {
	// Set the timestamps
	now := time.Now()
	user.CreatedAt = now
	user.LastUpdated = now

	// Convert the user to a DynamoDB item
	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String(UsersTableName),
		Item:      item,
		// Ensure the username doesn't already exist
		ConditionExpression: aws.String("attribute_not_exists(username)"),
	}

	// Put the item in DynamoDB
	_, err = r.db.Client.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser gets a user by username
func (r *UserRepository) GetUser(username string) (*models.User, error) {
	// Create the input for the GetItem operation
	input := &dynamodb.GetItemInput{
		TableName: aws.String(UsersTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
	}

	// Get the item from DynamoDB
	result, err := r.db.Client.GetItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if the item exists
	if result.Item == nil {
		return nil, nil
	}

	// Unmarshal the item into a User struct
	var user models.User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// ListUsers gets all users
func (r *UserRepository) ListUsers() ([]models.UserResponse, error) {
	// Create the input for the Scan operation
	input := &dynamodb.ScanInput{
		TableName: aws.String(UsersTableName),
	}

	// Scan the table
	result, err := r.db.Client.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Unmarshal the items into User structs
	var users []models.User
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}

	// Convert to UserResponse structs
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	return userResponses, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(user models.User) error {
	// Set the last updated timestamp
	user.LastUpdated = time.Now()

	// Convert the user to a DynamoDB item
	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String(UsersTableName),
		Item:      item,
		// Ensure the username exists
		ConditionExpression: aws.String("attribute_exists(username)"),
	}

	// Put the item in DynamoDB
	_, err = r.db.Client.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user by username
func (r *UserRepository) DeleteUser(username string) error {
	// Create the input for the DeleteItem operation
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(UsersTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
	}

	// Delete the item from DynamoDB
	_, err := r.db.Client.DeleteItem(input)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
