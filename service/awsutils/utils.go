package awsutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/mcastellin/aws-fail-az/domain"
	"golang.org/x/exp/slices"
)

func TokenizeResourceFilter(filter string, validKeys []string) (map[string]string, error) {
	filters := map[string]string{}

	if filter != "" {
		for _, attr := range strings.Split(filter, ";") {
			if attr != "" {
				tokens := strings.Split(attr, "=")
				if len(tokens) != 2 {
					err := fmt.Errorf(
						"Could not parse filter attribute. Expected format `key=value`, found %s",
						attr,
					)
					return map[string]string{}, err
				}

				key, value := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
				if key == "" || value == "" {
					err := fmt.Errorf("Could not parse filter attribute. Found empty key or value: %s", attr)
					return map[string]string{}, err
				} else if !slices.Contains(validKeys, key) {
					err := fmt.Errorf("Could not parse filter. Found unrecognized key `%s`", key)
					return map[string]string{}, err
				}

				filters[key] = value
			}
		}
	}

	return filters, nil
}

// Filter a list of subnets by Availability Zone
// Returns all subnets in the `subnetIds` list that are not attached to one of the availability
// zones in the `azs` parameter
func FilterSubnetsNotInAzs(api domain.Ec2Api, subnetIds []string, azs []string) ([]string, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: subnetIds,
	}
	describeSubnetsOutput, err := api.DescribeSubnets(context.TODO(), input)
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
