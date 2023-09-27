package service

import (
	"fmt"

	"github.com/mcastellin/aws-fail-az/awsapis"
	"github.com/mcastellin/aws-fail-az/domain"
	"github.com/mcastellin/aws-fail-az/service/asg"
	"github.com/mcastellin/aws-fail-az/service/ecs"
	"github.com/mcastellin/aws-fail-az/service/elbv2"
	"github.com/mcastellin/aws-fail-az/state"
)

// Initialize functions to create and recover faults for service type
func InitServiceFaults() *FaultsInitFns {

	initFns := &FaultsInitFns{
		faults: map[string]func(domain.TargetSelector, awsapis.AWSProvider) ([]domain.ConsistentStateResource, error){

			// Register init functions for new fault types in this structure

			domain.ResourceTypeEcsService:        ecs.NewEcsServiceFaultFromConfig,
			domain.ResourceTypeAutoScalingGroup:  asg.NewAutoScalingGroupFaultFromConfig,
			domain.ResourceTypeElbv2LoadBalancer: elbv2.NewElbv2LoadBalancerFaultFromConfig,
		},

		restore: map[string]func([]byte, awsapis.AWSProvider) error{

			// Register restore from state functions for new fault types in this structure

			domain.ResourceTypeEcsService:        ecs.RestoreEcsServicesFromState,
			domain.ResourceTypeAutoScalingGroup:  asg.RestoreAutoScalingGroupsFromState,
			domain.ResourceTypeElbv2LoadBalancer: elbv2.RestoreElbv2LoadBalancersFromState,
		},
	}
	return initFns
}

// An object that maps faults resource types to their init and restore functions
type FaultsInitFns struct {

	// A map of all available fault types and their initialization functions
	faults map[string]func(domain.TargetSelector, awsapis.AWSProvider) ([]domain.ConsistentStateResource, error)

	// A map of all available fault types and their restore functions
	restore map[string]func([]byte, awsapis.AWSProvider) error
}

// Initialize new resource faults from their selector
func (obj *FaultsInitFns) NewResourceForType(selector domain.TargetSelector,
	provider awsapis.AWSProvider) ([]domain.ConsistentStateResource, error) {

	initFn, ok := obj.faults[selector.Type]
	if ok {
		return initFn(selector, provider)
	}

	err := fmt.Errorf("Could not recognize resource type %s", selector.Type)
	return nil, err
}

// Call the specific RestoreFromState function for the resource type specified in the state object
func (obj *FaultsInitFns) RestoreFromState(state state.ResourceState, provider awsapis.AWSProvider) error {
	restoreFn, ok := obj.restore[state.ResourceType]
	if ok {
		return restoreFn(state.State, provider)
	}

	err := fmt.Errorf("unknown resource of type %s found in state with key %s. Object will be ignored",
		state.ResourceType,
		state.Key,
	)
	return err
}
