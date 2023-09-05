package awsapis

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Interfaces
type DynamodbApi interface {
	DynamodbItemGetter
	DynamodbItemPutter
	DynamodbItemDeleter
	DynamodbTableDescriptor
	DynamodbTableCreator
	DynamodbQueryPaginator
	DynamodbTableExistsWaiterIface
}

type DynamodbItemGetter interface {
	GetItem(ctx context.Context,
		params *dynamodb.GetItemInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

type DynamodbItemPutter interface {
	PutItem(ctx context.Context,
		params *dynamodb.PutItemInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type DynamodbItemDeleter interface {
	DeleteItem(ctx context.Context,
		params *dynamodb.DeleteItemInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

type DynamodbTableDescriptor interface {
	DescribeTable(ctx context.Context,
		params *dynamodb.DescribeTableInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
}

type DynamodbTableCreator interface {
	CreateTable(ctx context.Context,
		params *dynamodb.CreateTableInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
}

type DynamodbQueryPaginator interface {
	NewQueryPaginator(params *dynamodb.QueryInput) DynamodbQueryPager
}

type DynamodbQueryPager interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

type DynamodbTableExistsWaiterIface interface {
	NewTableExistsWaiter() DynamodbTableExistsWaiter
}

type DynamodbTableExistsWaiter interface {
	Wait(ctx context.Context, params *dynamodb.DescribeTableInput,
		maxWaitDur time.Duration, optFns ...func(*dynamodb.TableExistsWaiterOptions)) error
}

// Implementation
type AwsDynamodbApi struct {
	client *dynamodb.Client
}

func (a *AwsDynamodbApi) GetItem(ctx context.Context,
	params *dynamodb.GetItemInput,
	optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return a.client.GetItem(ctx, params, optFns...)
}

func (a *AwsDynamodbApi) PutItem(ctx context.Context,
	params *dynamodb.PutItemInput,
	optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return a.client.PutItem(ctx, params, optFns...)
}

func (a *AwsDynamodbApi) DescribeTable(ctx context.Context,
	params *dynamodb.DescribeTableInput,
	optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	return a.client.DescribeTable(ctx, params, optFns...)
}

func (a *AwsDynamodbApi) CreateTable(ctx context.Context,
	params *dynamodb.CreateTableInput,
	optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
	return a.client.CreateTable(ctx, params, optFns...)
}

func (a *AwsDynamodbApi) DeleteItem(ctx context.Context,
	params *dynamodb.DeleteItemInput,
	optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return a.client.DeleteItem(ctx, params, optFns...)
}

func (a *AwsDynamodbApi) NewQueryPaginator(params *dynamodb.QueryInput) DynamodbQueryPager {
	return dynamodb.NewQueryPaginator(a.client, params)
}

func (a *AwsDynamodbApi) NewTableExistsWaiter() DynamodbTableExistsWaiter {
	return dynamodb.NewTableExistsWaiter(a.client)
}
