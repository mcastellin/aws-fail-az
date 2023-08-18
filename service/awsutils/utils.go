package awsutils

// Filter a list of subnets by Availability Zone

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"golang.org/x/exp/slices"
)

// Returns all subnets in the `subnetIds` list that are not attached to one of the availability
// zones in the `azs` parameter
func FilterSubnetsNotInAzs(client *ec2.Client, subnetIds []string, azs []string) ([]string, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: subnetIds,
	}
	describeSubnetsOutput, err := client.DescribeSubnets(context.TODO(), input)
	if err != nil {
		return []string{}, err
	}

	newSubnets := []string{}
	for _, subnet := range describeSubnetsOutput.Subnets {
		if !slices.Contains(azs, *subnet.AvailabilityZone) {
			newSubnets = append(newSubnets, *subnet.SubnetId)
		}
	}

	return newSubnets, nil
}
