package db

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/devreserve/server/config"
)

// DynamoDBClient represents a client for interacting with DynamoDB
type DynamoDBClient struct {
	Client *dynamodb.DynamoDB
	Config config.Config
}

// DynamoDB table names
const (
	UsersTableName        = "DevReserve_Users"
	EnvironmentsTableName = "DevReserve_Environments"
	ReservationsTableName = "DevReserve_Reservations"
)

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(cfg config.Config) (*DynamoDBClient, error) {
	// Configure AWS session
	awsConfig := &aws.Config{
		Region: aws.String(cfg.AWSRegion),
	}

	// If a local endpoint is configured (for local development), use it
	if cfg.DynamoDBEndpoint != "" {
		awsConfig.Endpoint = aws.String(cfg.DynamoDBEndpoint)
		// For local development, we can use dummy credentials
		awsConfig.Credentials = credentials.NewStaticCredentials("dummy", "dummy", "")
	} else {
		// Check for explicit AWS credentials from environment variables
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKey != "" && secretKey != "" {
			awsConfig.Credentials = credentials.NewStaticCredentials(
				accessKey,
				secretKey,
				"", // token can be empty for regular access keys
			)
		}
	}

	// Create a new AWS session
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create a new DynamoDB client
	dbClient := dynamodb.New(sess)

	return &DynamoDBClient{
		Client: dbClient,
		Config: cfg,
	}, nil
}

// CreateTablesIfNotExist ensures that all required DynamoDB tables exist
func (db *DynamoDBClient) CreateTablesIfNotExist() error {
	// Create Users table if it doesn't exist
	if err := db.createUsersTable(); err != nil {
		return err
	}

	// Create Environments table if it doesn't exist
	if err := db.createEnvironmentsTable(); err != nil {
		return err
	}

	// Create Reservations table if it doesn't exist
	if err := db.createReservationsTable(); err != nil {
		return err
	}

	log.Println("All DynamoDB tables have been created or already exist")
	return nil
}

// createUsersTable creates the Users table if it doesn't exist
func (db *DynamoDBClient) createUsersTable() error {
	exists, err := db.tableExists(UsersTableName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(UsersTableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("username"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("username"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err = db.Client.CreateTable(input)
	if err != nil {
		return fmt.Errorf("failed to create Users table: %w", err)
	}

	log.Println("Created Users table")
	return nil
}

// createEnvironmentsTable creates the Environments table if it doesn't exist
func (db *DynamoDBClient) createEnvironmentsTable() error {
	exists, err := db.tableExists(EnvironmentsTableName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(EnvironmentsTableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err = db.Client.CreateTable(input)
	if err != nil {
		return fmt.Errorf("failed to create Environments table: %w", err)
	}

	log.Println("Created Environments table")
	return nil
}

// createReservationsTable creates the Reservations table if it doesn't exist
func (db *DynamoDBClient) createReservationsTable() error {
	exists, err := db.tableExists(ReservationsTableName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(ReservationsTableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("environmentId"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("EnvironmentIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("environmentId"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err = db.Client.CreateTable(input)
	if err != nil {
		return fmt.Errorf("failed to create Reservations table: %w", err)
	}

	log.Println("Created Reservations table")
	return nil
}

// tableExists checks if a table exists in DynamoDB
func (db *DynamoDBClient) tableExists(tableName string) (bool, error) {
	input := &dynamodb.ListTablesInput{}
	result, err := db.Client.ListTables(input)
	if err != nil {
		return false, fmt.Errorf("failed to list tables: %w", err)
	}

	for _, name := range result.TableNames {
		if *name == tableName {
			return true, nil
		}
	}

	return false, nil
}
