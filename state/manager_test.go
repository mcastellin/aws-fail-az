package state

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mcastellin/aws-fail-az/mock_awsapis"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestStateInitializeNewTableWithOsVar(t *testing.T) {
	t.Setenv("AWS_FAIL_AZ_STATE_TABLE", "test-value")

	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := mock_awsapis.NewMockDynamodbApi(ctrl)

	validVersion := ResourceState{
		Namespace:    "_system",
		Key:          "/schema/version",
		ResourceKey:  STATE_TABLE_SCHEMA_VERSION,
		ResourceType: "nil",
	}

	item, err := attributevalue.MarshalMap(validVersion)
	assert.Nil(t, err)

	mockApi.EXPECT().GetItem(gomock.Any(), gomock.Any()).Times(1).
		Return(&dynamodb.GetItemOutput{Item: item}, nil)

	mockApi.EXPECT().DescribeTable(gomock.Any(), tableNameInputMatch{"test-value"}).
		Times(1).
		Return(&dynamodb.DescribeTableOutput{}, nil)

	mgr := StateManagerImpl{
		Api: mockApi,
	}

	mgr.Initialize()

	assert.True(t, mgr.isInitialized)
}

func TestStateInitializeNewTable(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := mock_awsapis.NewMockDynamodbApi(ctrl)
	mockWaiter := mock_awsapis.NewMockDynamodbTableExistsWaiter(ctrl)

	mockWaiter.EXPECT().Wait(gomock.Any(), tableNameInputMatch{FALLBACK_STATE_TABLE_NAME}, gomock.Any()).
		Times(1).
		Return(nil)

	validVersion := ResourceState{
		Namespace:    "_system",
		Key:          "/schema/version",
		ResourceKey:  STATE_TABLE_SCHEMA_VERSION,
		ResourceType: "nil",
	}

	item, err := attributevalue.MarshalMap(validVersion)
	assert.Nil(t, err)

	gomock.InOrder(
		mockApi.EXPECT().GetItem(gomock.Any(), gomock.Any()).Times(1).
			Return(&dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{}}, nil),
		mockApi.EXPECT().GetItem(gomock.Any(), gomock.Any()).Times(1).
			Return(&dynamodb.GetItemOutput{Item: item}, nil),
	)

	mockApi.EXPECT().PutItem(gomock.Any(), gomock.Any()).Times(1).
		Return(&dynamodb.PutItemOutput{}, nil)

	mockApi.EXPECT().NewTableExistsWaiter().Times(1).Return(mockWaiter)
	mockApi.EXPECT().DescribeTable(gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil, &types.ResourceNotFoundException{})

	mockApi.EXPECT().CreateTable(gomock.Any(), gomock.Any()).
		Times(1)

	mgr := StateManagerImpl{Api: mockApi}

	mgr.Initialize()
}

func TestSaveStateShouldNotOverrideExistingKeys(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	mockApi := mock_awsapis.NewMockDynamodbApi(ctrl)

	mockApi.EXPECT().GetItem(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, params *dynamodb.GetItemInput,
			f ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {

			item, _ := attributevalue.MarshalMap(ResourceState{Key: "dummy"})
			return &dynamodb.GetItemOutput{Item: item}, nil
		})

	mgr := StateManagerImpl{Api: mockApi}
	mgr.isInitialized = true

	err := mgr.Save("type", "key", []byte("payload"))

	assert.NotNil(t, err)
}

// Matchers
func TableNameInputMatcher(x string) gomock.Matcher { return tableNameInputMatch{x} }

type tableNameInputMatch struct {
	x string
}

func (m tableNameInputMatch) Matches(x interface{}) bool {
	if y, ok := x.(*dynamodb.DescribeTableInput); ok {
		return m.x == *y.TableName
	}
	return false
}

func (m tableNameInputMatch) String() string {
	return m.x
}
