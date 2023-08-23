package state

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mcastellin/aws-fail-az/awsapis"
)

// The default table name to store resource states
const FALLBACK_STATE_TABLE_NAME string = "aws-fail-az-state"

type StateManager interface {
	Initialize()
	Save(resourceType string, resourceKey string, state []byte) error
	ReadStates() ([]ResourceState, error)
	RemoveState(stateObj ResourceState) error
}

type ResourceState struct {
	Namespace    string `dynamodbav:"namespace"`
	Key          string `dynamodbav:"key"`
	ResourceKey  string `dynamodbav:"resourceKey"`
	ResourceType string `dynamodbav:"resourceType"`
	CreatedTime  int64  `dynamodbav:"createdTime"`
	State        []byte `dynamodbav:"state"`
}

func (state ResourceState) GetKey() map[string]types.AttributeValue {
	namespace, err := attributevalue.Marshal(state.Namespace)
	if err != nil {
		log.Panic(err)
	}
	key, err := attributevalue.Marshal(state.Key)
	if err != nil {
		log.Panic(err)
	}
	return map[string]types.AttributeValue{"namespace": namespace, "key": key}
}

// A State Manager object to interact with resource state
type StateManagerImpl struct {
	Api       awsapis.DynamodbApi
	TableName string
	Namespace string
}

// Check if the state table already exists for the current AWS Account/Region
// Returns: true if the table exists, false otherwise
func (m StateManagerImpl) tableExists() (bool, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(m.TableName),
	}

	_, err := m.Api.DescribeTable(context.TODO(), input)
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
func (m StateManagerImpl) createTable() (*dynamodb.CreateTableOutput, error) {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(m.TableName),
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
				IndexName: aws.String("LSINamespace"),
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

	createOutput, err := m.Api.CreateTable(context.TODO(), input)
	if err != nil {
		log.Fatalf("Failed to create Dynamodb Table to store the current resource state, %v", err)
	}

	log.Printf("Wait for table exists: %s", m.TableName)
	waiter := m.Api.NewTableExistsWaiter()
	err = waiter.Wait(
		context.TODO(),
		&dynamodb.DescribeTableInput{TableName: aws.String(m.TableName)},
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
func (m *StateManagerImpl) Initialize() {
	stateTableName := os.Getenv("AWS_FAILAZ_STATE_TABLE")
	if stateTableName == "" {
		log.Printf("AWS_FAILAZ_STATE_STABLE variable is not set. Using default %s", FALLBACK_STATE_TABLE_NAME)
		m.TableName = FALLBACK_STATE_TABLE_NAME
	} else {
		m.TableName = stateTableName
	}

	exists, err := m.tableExists()
	if err != nil {
		log.Fatalf("An unknown error occurred: %v", err)
	}

	if !exists {
		log.Printf("State table with name %s not found. Creating...", stateTableName)
		_, err := m.createTable()
		if err != nil {
			log.Fatalf("Error creating state table in Dynamodb. %v", err)
		}
	}

	if m.Namespace == "" {
		m.Namespace = "default"
	}
}

func (m StateManagerImpl) Save(resourceType string, resourceKey string, state []byte) error {

	key := fmt.Sprintf("/%s/%s/%s", m.Namespace, resourceType, resourceKey)
	stateObj := ResourceState{
		Namespace:    m.Namespace,
		Key:          key,
		ResourceKey:  resourceKey,
		ResourceType: resourceType,
		CreatedTime:  time.Now().Unix(),
		State:        state,
	}

	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(m.TableName),
		Key:       stateObj.GetKey(),
	}
	response, err := m.Api.GetItem(context.TODO(), getItemInput)
	if err != nil {
		return err
	}
	keyExists := len(response.Item) > 0
	if keyExists {
		return fmt.Errorf("State key already exist for resource %s", key)
	}

	item, err := attributevalue.MarshalMap(stateObj)
	if err != nil {
		return err
	}
	_, err = m.Api.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(m.TableName),
		Item:      item,
	})
	if err != nil {
		return err
	}

	return nil
}

func (m StateManagerImpl) ReadStates() ([]ResourceState, error) {

	keyExpr := expression.Key("namespace").Equal(expression.Value(m.Namespace))
	expr, err := expression.NewBuilder().WithKeyCondition(keyExpr).Build()
	if err != nil {
		log.Println("Unable to build query expression to fetch resource states")
		return []ResourceState{}, err
	}

	resourceStates := []ResourceState{}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(m.TableName),
		IndexName:                 aws.String("LSINamespace"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	paginator := m.Api.NewQueryPaginator(queryInput)
	for paginator.HasMorePages() {
		queryOutput, err := paginator.NextPage(context.TODO())
		if err != nil {
			return []ResourceState{}, err
		}

		var states []ResourceState
		err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &states)
		if err != nil {
			log.Println("Error unmarshalling resource states")
			return []ResourceState{}, err
		} else {
			resourceStates = append(resourceStates, states...)
		}
	}

	return resourceStates, nil
}

func (m StateManagerImpl) RemoveState(stateObj ResourceState) error {
	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(m.TableName),
		Key:       stateObj.GetKey(),
	}
	_, err := m.Api.DeleteItem(context.TODO(), deleteItemInput)
	if err != nil {
		return err
	}

	return nil
}
