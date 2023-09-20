package ecs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/mcastellin/aws-fail-az/awsapis_mocks"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFilterServiceByTagsShouldExcludeResults(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	listClustersPager := createListClusterPager(ctrl, [][]string{{"test-cluster"}})
	listServicesPager := createListServicesPager(ctrl, [][]string{{"test-service"}})

	mockEcsAPI := awsapis_mocks.NewMockEcsApi(ctrl)
	mockEcsAPI.EXPECT().NewListClustersPaginator(gomock.Any()).Times(1).Return(listClustersPager)
	mockEcsAPI.EXPECT().NewListServicesPaginator(gomock.Any()).Times(1).Return(listServicesPager)

	mockEcsAPI.EXPECT().ListTagsForResource(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
		DoAndReturn(func(_ context.Context, param *ecs.ListTagsForResourceInput, optFns ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error) {
			return &ecs.ListTagsForResourceOutput{
				Tags: []types.Tag{{Key: aws.String("Application"), Value: aws.String("live-app")}},
			}, nil
		})

	mockProvider := awsapis_mocks.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewEcsApi().AnyTimes().Return(mockEcsAPI)
	mockProvider.EXPECT().NewEcsApi().AnyTimes().Return(mockEcsAPI)

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

	listClustersPager := createListClusterPager(ctrl, [][]string{{"test-cluster"}})
	listServicesPager := createListServicesPager(ctrl, [][]string{{"test-service"}})

	mockEcsAPI := awsapis_mocks.NewMockEcsApi(ctrl)
	mockEcsAPI.EXPECT().NewListClustersPaginator(gomock.Any()).Times(1).Return(listClustersPager)
	mockEcsAPI.EXPECT().NewListServicesPaginator(gomock.Any()).Times(1).Return(listServicesPager)

	mockEcsAPI.EXPECT().ListTagsForResource(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
		DoAndReturn(func(_ context.Context, param *ecs.ListTagsForResourceInput, optFns ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error) {
			return &ecs.ListTagsForResourceOutput{
				Tags: []types.Tag{{Key: aws.String("Application"), Value: aws.String("live-app")}},
			}, nil
		})

	mockProvider := awsapis_mocks.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewEcsApi().AnyTimes().Return(mockEcsAPI)
	mockProvider.EXPECT().NewEcsApi().AnyTimes().Return(mockEcsAPI)

	config := domain.TargetSelector{
		Type: RESOURCE_TYPE,
		Tags: []domain.AWSTag{{Name: "Application", Value: "live-app"}},
	}

	results, err := NewFromConfig(config, mockProvider)

	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-service", results[0].(*ECSService).ServiceName)

}

func TestFilterServiceByTagsShouldMatchResultsFromAllPages(t *testing.T) {
	ctrl, _ := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	listClustersPager := createListClusterPager(ctrl, [][]string{{"test-cluster"}})
	listServicesPager := createListServicesPager(ctrl, [][]string{{"test-service", "test-service-2"}, {"test-service-3"}})

	mockEcsAPI := awsapis_mocks.NewMockEcsApi(ctrl)
	mockEcsAPI.EXPECT().NewListClustersPaginator(gomock.Any()).Times(1).Return(listClustersPager)
	mockEcsAPI.EXPECT().NewListServicesPaginator(gomock.Any()).Times(1).Return(listServicesPager)

	mockEcsAPI.EXPECT().ListTagsForResource(gomock.Any(), gomock.Any(), gomock.Any()).Times(3).
		DoAndReturn(func(_ context.Context, param *ecs.ListTagsForResourceInput, optFns ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error) {
			return &ecs.ListTagsForResourceOutput{
				Tags: []types.Tag{{Key: aws.String("Application"), Value: aws.String("live-app")}},
			}, nil
		})

	mockProvider := awsapis_mocks.NewMockAWSProvider(ctrl)
	mockProvider.EXPECT().NewEcsApi().AnyTimes().Return(mockEcsAPI)
	mockProvider.EXPECT().NewEcsApi().AnyTimes().Return(mockEcsAPI)

	config := domain.TargetSelector{
		Type: RESOURCE_TYPE,
		Tags: []domain.AWSTag{{Name: "Application", Value: "live-app"}},
	}

	results, err := NewFromConfig(config, mockProvider)

	assert.Nil(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, "test-service", results[0].(*ECSService).ServiceName)
	assert.Equal(t, "test-service-2", results[1].(*ECSService).ServiceName)
	assert.Equal(t, "test-service-3", results[2].(*ECSService).ServiceName)

}

func createListClusterPager(ctrl *gomock.Controller, arnsPages [][]string) *awsapis_mocks.MockListClustersPager {
	mockListClusterPager := awsapis_mocks.NewMockListClustersPager(ctrl)
	gomock.InOrder(
		mockListClusterPager.EXPECT().HasMorePages().Times(len(arnsPages)).Return(true),
		mockListClusterPager.EXPECT().HasMorePages().Times(1).Return(false),
	)
	calls := []*gomock.Call{}
	for idx := range arnsPages {
		c := mockListClusterPager.EXPECT().NextPage(gomock.Any()).Times(1).
			Return(&ecs.ListClustersOutput{
				ClusterArns: arnsPages[idx],
			}, nil)
		calls = append(calls, c)
	}
	gomock.InOrder(calls...)

	return mockListClusterPager
}

func createListServicesPager(ctrl *gomock.Controller, arnsPages [][]string) *awsapis_mocks.MockListServicesPager {
	mockListServicePager := awsapis_mocks.NewMockListServicesPager(ctrl)
	gomock.InOrder(
		mockListServicePager.EXPECT().HasMorePages().Times(len(arnsPages)).Return(true),
		mockListServicePager.EXPECT().HasMorePages().Times(1).Return(false),
	)
	calls := []*gomock.Call{}
	for idx := range arnsPages {
		c := mockListServicePager.EXPECT().NextPage(gomock.Any()).Times(1).
			Return(&ecs.ListServicesOutput{
				ServiceArns: arnsPages[idx],
			}, nil)
		calls = append(calls, c)
	}
	gomock.InOrder(calls...)

	return mockListServicePager
}
