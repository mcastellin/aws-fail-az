package domain

import "github.com/mcastellin/aws-fail-az/state"

// A representation of an AWS service state that can be
// validated and stored with StateManager
type ConsistentServiceState interface {
	Check() (bool, error)
	Save(*state.StateManager) error
	Fail([]string) error
	Restore([]byte) error
}

// AZ Failure Configuration
type FaultConfiguration struct {
	Azs      []string          `json:"azs"`
	Services []ServiceSelector `json:"services"`
}

// AWS ServiceSelector
type ServiceSelector struct {
	Type   string   `json:"type"`
	Filter string   `json:"filter"`
	Tags   []AWSTag `json:"tags"`
}

// AWS Tag
type AWSTag struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}
