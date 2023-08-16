package state

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// The default table name to store resource states
const fallbackStateTableName string = "aws-fail-az-state"

// A State Manager object to interact with resource state
type StateManager struct {
	Client    *dynamodb.Client
	TableName string
	Namespace string
}

// Check if the state table already exists for the current AWS Account/Region
// Returns: true if the table exists, false otherwise
func (manager StateManager) tableExists() (bool, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(manager.TableName),
	}

	_, err := manager.Client.DescribeTable(context.TODO(), input)
	if err != nil {
		if t := new(types.ResourceNotFoundException); errors.As(err, &t) {
			return false, nil // Table does not exists
		}
		return false, err // Other error occurred
	}

	return true, nil // Table exists
}

// Creates the resource state table in Dynamodb for the current AWS Account/Region
// and wait for table creationg before returning
func (manager StateManager) createTable() (*dynamodb.CreateTableOutput, error) {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(manager.TableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("namespace"),
				KeyType:       types.KeyTypeHash,
			}, {
				AttributeName: aws.String("key"),
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("namespace"),
				AttributeType: types.ScalarAttributeTypeS,
			}, {
				AttributeName: aws.String("key"),
				AttributeType: types.ScalarAttributeTypeS,
			}, {
				AttributeName: aws.String("createdTime"),
				AttributeType: types.ScalarAttributeTypeN,
			},
		},
		LocalSecondaryIndexes: []types.LocalSecondaryIndex{
			{
				IndexName: aws.String("LSITestId"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("namespace"),
						KeyType:       types.KeyTypeHash,
					}, {
						AttributeName: aws.String("createdTime"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	}

	createOutput, err := manager.Client.CreateTable(context.TODO(), input)
	if err != nil {
		log.Fatalf("Failed to create Dynamodb Table to store the current resource state, %v", err)
	}

	log.Printf("Wait for table exists: %s", manager.TableName)
	waiter := dynamodb.NewTableExistsWaiter(manager.Client)
	err = waiter.Wait(
		context.TODO(),
		&dynamodb.DescribeTableInput{TableName: aws.String(manager.TableName)},
		5*time.Minute,
	)
	if err != nil {
		log.Fatalf("Wait for table exists failed. It's not safe to continue this operation. %v", err)
	}

	return createOutput, nil
}

// Initialize the state manager.
// This only needs to be called once at the beginning of the program to create the
// state table in Dynamodb. Further calls will have no effect.
func (manager StateManager) Initialize() {
	stateTableName := os.Getenv("AWS_FAILAZ_STATE_TABLE")
	if stateTableName == "" {
		log.Printf("AWS_FAILAZ_STATE_STABLE variable is not set. Using default %s", fallbackStateTableName)
		manager.TableName = fallbackStateTableName
	} else {
		manager.TableName = stateTableName
	}

	exists, err := manager.tableExists()
	if err != nil {
		log.Fatalf("An unknown error occurred: %v", err)
	}

	if !exists {
		log.Printf("State table with name %s not found. Creating...", stateTableName)
		_, err := manager.createTable()
		if err != nil {
			log.Fatalf("Error creating state table in Dynamodb. %v", err)
		}
	}
}
