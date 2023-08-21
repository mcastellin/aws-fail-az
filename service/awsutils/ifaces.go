package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type EC2DescribeSubnetsAPI interface {
	DescribeSubnets(ctx context.Context,
		params *ec2.DescribeSubnetsInput,
		optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
}
