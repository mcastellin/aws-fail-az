package state

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/mock_awsapis"
	"go.uber.org/mock/gomock"
)

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

func TestStateInitializeNewTable(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	mockApi := mock_awsapis.NewMockDynamodbApi(ctrl)
	mockWaiter := mock_awsapis.NewMockDynamodbTableExistsWaiter(ctrl)

	mockWaiter.EXPECT().Wait(gomock.Any(), tableNameInputMatch{FALLBACK_STATE_TABLE_NAME}, gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, params *dynamodb.DescribeTableInput,
			maxWaitDur time.Duration, optFns ...func(*dynamodb.TableExistsWaiterOptions)) error {
			return nil
		})

	mockApi.EXPECT().NewTableExistsWaiter().
		Times(1).
		DoAndReturn(func() awsapis.DynamodbTableExistsWaiter {
			return mockWaiter
		})
	mockApi.EXPECT().DescribeTable(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, params *dynamodb.DescribeTableInput,
			f ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
			err := &types.ResourceNotFoundException{}
			return nil, err
		})

	mockApi.EXPECT().CreateTable(gomock.Any(), gomock.Any()).
		Times(1)

	mgr := StateManagerImpl{
		Api: mockApi,
	}

	mgr.Initialize()
}
