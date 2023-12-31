package elbv2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/awsutils"
	"github.com/mcastellin/aws-fail-az/state"
)

type LoadBalancerState struct {
	LoadBalancerName string   `json:"lbName"`
	Subnets          []string `json:"subnets"`
}
type LoadBalancer struct {
	Provider awsapis.AWSProvider
	Name     string

	stateSubnets []string
}

func (lb *LoadBalancer) Check() (bool, error) {
	log.Printf("%s name=%s: checking resource state before failure simulation",
		domain.ResourceTypeElbv2LoadBalancer, lb.Name)

	api := lb.Provider.NewElbV2Api()

	output, err := describeLoadBalancer(api, lb.Name)
	if err != nil {
		return false, err
	}
	if len(output.LoadBalancers) == 0 {
		return false, fmt.Errorf("Could not describe load balancer with name %s", lb.Name)
	}
	subnetIds := getLoadBalancerSubnets(output.LoadBalancers[0])

	if len(subnetIds) <= 2 {
		return false, fmt.Errorf("Insufficient number of subnets for resource %s."+
			" Load balancers require a minimum of 3 availability zones to simulate AZ failure, found %d.",
			lb.Name, len(subnetIds))
	}

	return true, nil
}

func (lb *LoadBalancer) Save(stateManager state.StateManager) error {

	api := lb.Provider.NewElbV2Api()

	describeOutput, err := describeLoadBalancer(api, lb.Name)
	if err != nil {
		return err
	}
	if len(describeOutput.LoadBalancers) == 0 {
		return fmt.Errorf("Could not describe load balancer with name %s", lb.Name)
	}
	loadBalancerDescriptor := describeOutput.LoadBalancers[0]
	subnetIds := getLoadBalancerSubnets(loadBalancerDescriptor)

	state := &LoadBalancerState{
		LoadBalancerName: lb.Name,
		Subnets:          subnetIds,
	}
	data, err := json.Marshal(state)
	if err != nil {
		log.Println("Error while marshalling load balancer state")
		return err
	}
	err = stateManager.Save(domain.ResourceTypeElbv2LoadBalancer, lb.Name, data)

	return err
}

func (lb *LoadBalancer) Fail(azs []string) error {

	api := lb.Provider.NewElbV2Api()
	ec2Api := lb.Provider.NewEc2Api()

	describeOutput, err := describeLoadBalancer(api, lb.Name)
	if err != nil {
		return err
	}
	if len(describeOutput.LoadBalancers) == 0 {
		return fmt.Errorf("Could not describe load balancer with name %s", lb.Name)
	}
	loadBalancerDescriptor := describeOutput.LoadBalancers[0]
	subnetIds := getLoadBalancerSubnets(loadBalancerDescriptor)

	newSubnets, err := awsutils.FilterSubnetsNotInAzs(ec2Api, subnetIds, azs)
	if err != nil {
		log.Printf("Error while filtering subnets by AZs: %v", err)
		return err
	}
	if len(newSubnets) <= 1 {
		return fmt.Errorf("AZ failure for load-balancer %s would remove all but one subnets."+
			" Load balancers require at least 2 availability zones. AZ failure will now stop", lb.Name)
	}

	log.Printf("%s name=%s: failing AZs %s for load-balancer", domain.ResourceTypeElbv2LoadBalancer, lb.Name, azs)

	_, err = api.SetSubnets(context.TODO(), &elasticloadbalancingv2.SetSubnetsInput{
		LoadBalancerArn: loadBalancerDescriptor.LoadBalancerArn,
		Subnets:         newSubnets,
	})

	return err
}

func (lb *LoadBalancer) Restore() error {

	log.Printf("%s name=%s: restoring AZs for load-balancer", domain.ResourceTypeElbv2LoadBalancer, lb.Name)

	api := lb.Provider.NewElbV2Api()

	arn := aws.String(lb.Name)
	if !strings.HasPrefix(*arn, "arn:") {
		out, err := describeLoadBalancer(api, lb.Name)
		if err != nil {
			return err
		}
		arn = out.LoadBalancers[0].LoadBalancerArn
	}

	_, err := api.SetSubnets(context.TODO(), &elasticloadbalancingv2.SetSubnetsInput{
		LoadBalancerArn: arn,
		Subnets:         lb.stateSubnets,
	})
	return err
}

func describeLoadBalancer(api awsapis.ElbV2LoadBalancersDescriptor, name string) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {

	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}
	if strings.HasPrefix(name, "arn:") {
		input.LoadBalancerArns = []string{name}
	} else {
		input.Names = []string{name}
	}

	return api.DescribeLoadBalancers(context.TODO(), input)
}

func getLoadBalancerSubnets(descriptor types.LoadBalancer) []string {
	azs := descriptor.AvailabilityZones
	subnets := make([]string, len(azs))
	for idx, az := range azs {
		subnets[idx] = *az.SubnetId
	}
	return subnets
}
