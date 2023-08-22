package domain

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func NewEcsApi(provider *AWSProvider) EcsApi {
	return &AwsEcsApi{
		client: ecs.NewFromConfig(provider.GetConnection()),
	}
}

// Interfaces
type EcsApi interface {
	EcsTagsLister
	EcsServiceDescriptor
	EcsServiceUpdater
	EcsTaskDescriptor
	EcsTaskStopper
	ListClustersPaginator
	ListServicesPaginator
	ListTasksPaginator
}

type EcsTagsLister interface {
	ListTagsForResource(ctx context.Context,
		params *ecs.ListTagsForResourceInput,
		optFns ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error)
}

type EcsServiceDescriptor interface {
	DescribeServices(ctx context.Context,
		params *ecs.DescribeServicesInput,
		optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error)
}

type EcsServiceUpdater interface {
	UpdateService(ctx context.Context,
		params *ecs.UpdateServiceInput,
		optFns ...func(*ecs.Options)) (*ecs.UpdateServiceOutput, error)
}

type EcsTaskStopper interface {
	StopTask(ctx context.Context,
		params *ecs.StopTaskInput,
		optFns ...func(*ecs.Options)) (*ecs.StopTaskOutput, error)
}

type EcsTaskDescriptor interface {
	DescribeTasks(ctx context.Context,
		params *ecs.DescribeTasksInput,
		optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error)
}

type ListClustersPaginator interface {
	NewListClustersPaginator(params *ecs.ListClustersInput) ListClustersPager
}

type ListClustersPager interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*ecs.Options)) (*ecs.ListClustersOutput, error)
}

type ListServicesPaginator interface {
	NewListServicesPaginator(params *ecs.ListServicesInput) ListServicesPager
}

type ListServicesPager interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*ecs.Options)) (*ecs.ListServicesOutput, error)
}

type ListTasksPaginator interface {
	NewListTasksPaginator(params *ecs.ListTasksInput) ListTasksPager
}

type ListTasksPager interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*ecs.Options)) (*ecs.ListTasksOutput, error)
}

// Implementation
type AwsEcsApi struct {
	client *ecs.Client
}

func (a *AwsEcsApi) ListTagsForResource(ctx context.Context,
	params *ecs.ListTagsForResourceInput,
	optFns ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error) {

	return a.client.ListTagsForResource(ctx, params, optFns...)
}

func (a *AwsEcsApi) DescribeServices(ctx context.Context,
	params *ecs.DescribeServicesInput,
	optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error) {

	return a.client.DescribeServices(ctx, params, optFns...)
}

func (a *AwsEcsApi) UpdateService(ctx context.Context,
	params *ecs.UpdateServiceInput,
	optFns ...func(*ecs.Options)) (*ecs.UpdateServiceOutput, error) {

	return a.client.UpdateService(ctx, params, optFns...)
}

func (a *AwsEcsApi) DescribeTasks(ctx context.Context,
	params *ecs.DescribeTasksInput,
	optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error) {

	return a.client.DescribeTasks(ctx, params, optFns...)
}

func (a *AwsEcsApi) StopTask(ctx context.Context,
	params *ecs.StopTaskInput,
	optFns ...func(*ecs.Options)) (*ecs.StopTaskOutput, error) {

	return a.client.StopTask(ctx, params, optFns...)
}

func (a *AwsEcsApi) NewListClustersPaginator(params *ecs.ListClustersInput) ListClustersPager {
	return ecs.NewListClustersPaginator(a.client, params)
}

func (a *AwsEcsApi) NewListServicesPaginator(params *ecs.ListServicesInput) ListServicesPager {
	return ecs.NewListServicesPaginator(a.client, params)
}

func (a *AwsEcsApi) NewListTasksPaginator(params *ecs.ListTasksInput) ListTasksPager {
	return ecs.NewListTasksPaginator(a.client, params)
}
