package ecs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/mock_awsapis"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFilterServiceByTagsShouldExcludeResults(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	params := MockInput{
		ListClusterArns:     []string{"test-cluster"},
		ListClustersPages:   1,
		ListServicesArn:     []string{"test-service"},
		ListServicesPages:   1,
		ListTagsForResource: []types.Tag{{Key: aws.String("Application"), Value: aws.String("live-app")}},
	}
	mockProvider := createProvider(ctrl, params)
	config := domain.TargetSelector{
		Type: RESOURCE_TYPE,
		Tags: []domain.AWSTag{{Name: "Application", Value: "notfound"}},
	}

	results, err := NewFromConfig(config, mockProvider)

	assert.Nil(t, err)
	assert.Len(t, results, 0)

}

func TestFilterServiceByTagsShouldMatch(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	params := MockInput{
		ListClusterArns:     []string{"test-cluster"},
		ListClustersPages:   1,
		ListServicesArn:     []string{"test-service"},
		ListServicesPages:   1,
		ListTagsForResource: []types.Tag{{Key: aws.String("Application"), Value: aws.String("live-app")}},
	}
	mockProvider := createProvider(ctrl, params)
	config := domain.TargetSelector{
		Type: RESOURCE_TYPE,
		Tags: []domain.AWSTag{{Name: "Application", Value: "live-app"}},
	}

	results, err := NewFromConfig(config, mockProvider)

	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-service", results[0].(ECSService).ServiceName)

}

type MockInput struct {
	ListClusterArns     []string
	ListClustersPages   int
	ListServicesArn     []string
	ListServicesPages   int
	ListTagsForResource []types.Tag
}

func createProvider(ctrl *gomock.Controller, mockParam MockInput) *mock_awsapis.MockAWSProvider {
	mockListClusterPager := mock_awsapis.NewMockListClustersPager(ctrl)
	gomock.InOrder(
		mockListClusterPager.EXPECT().HasMorePages().Times(mockParam.ListClustersPages).Return(true),
		mockListClusterPager.EXPECT().HasMorePages().Times(1).Return(false),
	)
	mockListClusterPager.EXPECT().NextPage(gomock.Any()).Times(1).
		DoAndReturn(func(_ context.Context, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
			return &ecs.ListClustersOutput{
				ClusterArns: mockParam.ListClusterArns,
			}, nil
		})

	mockListServicePager := mock_awsapis.NewMockListServicesPager(ctrl)
	gomock.InOrder(
		mockListServicePager.EXPECT().HasMorePages().Times(mockParam.ListServicesPages).Return(true),
		mockListServicePager.EXPECT().HasMorePages().Times(1).Return(false),
	)
	mockListServicePager.EXPECT().NextPage(gomock.Any()).Times(1).
		DoAndReturn(func(_ context.Context, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
			return &ecs.ListServicesOutput{
				ServiceArns: mockParam.ListServicesArn,
			}, nil
		})

	mockEcsApi := mock_awsapis.NewMockEcsApi(ctrl)
	mockEcsApi.EXPECT().NewListClustersPaginator(gomock.Any()).Times(1).Return(mockListClusterPager)
	mockEcsApi.EXPECT().NewListServicesPaginator(gomock.Any()).Times(1).Return(mockListServicePager)

	mockEcsApi.EXPECT().ListTagsForResource(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
		DoAndReturn(func(_ context.Context, param *ecs.ListTagsForResourceInput, optFns ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error) {
			return &ecs.ListTagsForResourceOutput{
				Tags: mockParam.ListTagsForResource,
			}, nil
		})

	mockProvider := mock_awsapis.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewEcsApi().AnyTimes().Return(mockEcsApi)
	mockProvider.EXPECT().NewEcsApi().AnyTimes().Return(mockEcsApi)

	return mockProvider
}
