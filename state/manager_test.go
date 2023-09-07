package state

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

	matcher := describeTableInputMatcher{&dynamodb.DescribeTableInput{
		TableName: aws.String("test-value"),
	}}
	mockApi.EXPECT().DescribeTable(gomock.Any(), matcher).
		Times(1).
		Return(&dynamodb.DescribeTableOutput{}, nil)

	mgr := StateManagerImpl{
		Api: mockApi,
	}

	err = mgr.Initialize()

	assert.Nil(t, err)
	assert.True(t, mgr.isInitialized)
}

func TestStateInitializeShouldFailWithWrongVersion(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := mock_awsapis.NewMockDynamodbApi(ctrl)

	validVersion := ResourceState{
		Namespace:    "_system",
		Key:          "/schema/version",
		ResourceKey:  "--incompatibile-schema-version--",
		ResourceType: "nil",
	}

	item, err := attributevalue.MarshalMap(validVersion)
	assert.Nil(t, err)

	mockApi.EXPECT().GetItem(gomock.Any(), gomock.Any()).Times(1).
		Return(&dynamodb.GetItemOutput{Item: item}, nil)

	mockApi.EXPECT().DescribeTable(gomock.Any(), gomock.Any()).
		Times(1).
		Return(&dynamodb.DescribeTableOutput{}, nil)

	mgr := StateManagerImpl{
		Api: mockApi,
	}

	err = mgr.Initialize()

	assert.NotNil(t, err)
	assert.False(t, mgr.isInitialized)
}

func TestStateInitializeNewTable(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := mock_awsapis.NewMockDynamodbApi(ctrl)
	mockWaiter := mock_awsapis.NewMockDynamodbTableExistsWaiter(ctrl)

	matcher := describeTableInputMatcher{&dynamodb.DescribeTableInput{
		TableName: aws.String(FALLBACK_STATE_TABLE_NAME),
	}}
	mockWaiter.EXPECT().Wait(gomock.Any(), matcher, gomock.Any()).
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

	err = mgr.Initialize()

	assert.Nil(t, err)
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

type describeTableInputMatcher struct {
	x *dynamodb.DescribeTableInput
}

func (m describeTableInputMatcher) Matches(x interface{}) bool {
	if y, ok := x.(*dynamodb.DescribeTableInput); ok {
		return reflect.DeepEqual(m.x, y)
	}
	return false
}

func (m describeTableInputMatcher) String() string {
	return fmt.Sprintf("%v", m.x)
}
