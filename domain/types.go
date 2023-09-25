package domain

import (
	"fmt"

	"github.com/mcastellin/aws-fail-az/state"
)

const (
	ResourceTypeEcsService        = "ecs-service"
	ResourceTypeAutoScalingGroup  = "auto-scaling-group"
	ResourceTypeElbv2LoadBalancer = "elbv2-load-balancer"
)

// A representation of an AWS resource state that can be
// validated and stored with StateManager
type ConsistentStateResource interface {
	Check() (bool, error)
	Save(state.StateManager) error
	Fail([]string) error
	Restore() error
}

// AZ Failure Configuration
type FaultConfiguration struct {
	Azs     []string         `json:"azs"`
	Targets []TargetSelector `json:"targets"`
}

// AWS Tag
type AWSTag struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

// A struct to represent the selection of AWS resource targets
type TargetSelector struct {
	Type   string   `json:"type"`
	Filter string   `json:"filter"`
	Tags   []AWSTag `json:"tags"`
}

// Validates all required fields for target selector have been provided
func (t TargetSelector) Validate() error {
	if t.Filter != "" && len(t.Tags) > 0 {
		return fmt.Errorf("validation failed: Both 'filter' and 'tags' selectors specified. Only one allowed")
	}
	if t.Filter == "" && len(t.Tags) == 0 {
		return fmt.Errorf("validation failed: One of 'filter' and 'tags' selectors must be specified")
	}
	return nil
}
