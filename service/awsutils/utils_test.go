package awsutils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

type mockEC2DescribeSubnetsImpl struct{}

func (m mockEC2DescribeSubnetsImpl) DescribeSubnets(ctx context.Context,
	params *ec2.DescribeSubnetsInput,
	optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {

	output := &ec2.DescribeSubnetsOutput{
		Subnets: []types.Subnet{
			{
				SubnetId:         aws.String("s-1234"),
				AvailabilityZone: aws.String("us-east-1b"),
			},
			{
				SubnetId:         aws.String("s-0000"),
				AvailabilityZone: aws.String("us-east-1a"),
			},
		},
	}

	return output, nil
}

func TestFilterSubnetsNotInAzs(t *testing.T) {

	api := &mockEC2DescribeSubnetsImpl{}

	newSubnets, err := FilterSubnetsNotInAzs(api, []string{"s-1234", "s-0000"}, []string{"us-east-1b"})

	assert.Nil(t, err, "No errros should have been returned")
	assert.Equal(t, []string{"s-0000"}, newSubnets, "Should have returned only subnet not in failing az")

}

func TestTokenizeResourceFilter(t *testing.T) {
	attributes, err := TokenizeResourceFilter("cluster=test;service=test-service", []string{"cluster", "service"})

	expected := map[string]string{"cluster": "test", "service": "test-service"}
	assert.Nil(t, err)
	assert.Equal(t, expected, attributes)
}

func TestTokenizeResourceFilterShouldEliminateEmpty(t *testing.T) {
	attributes, err := TokenizeResourceFilter(";cluster=test;service=test-service;;", []string{"cluster", "service"})

	expected := map[string]string{"cluster": "test", "service": "test-service"}
	assert.Nil(t, err)
	assert.Equal(t, expected, attributes)
}

func TestTokenizeResourceFilterShouldTrimSpaces(t *testing.T) {
	attributes, err := TokenizeResourceFilter(";cluster  =   test;service = test service;;", []string{"cluster", "service"})

	expected := map[string]string{"cluster": "test", "service": "test service"}
	assert.Nil(t, err)
	assert.Equal(t, expected, attributes)
}

func TestTokenizeResourceFilterShouldRefuseInvalidKeys(t *testing.T) {
	_, err := TokenizeResourceFilter(";cluster=test;service=test-service;;", []string{"cluster", "another"})

	assert.NotNil(t, err)
}
