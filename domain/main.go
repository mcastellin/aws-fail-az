package domain

import (
	"fmt"

	"github.com/mcastellin/aws-fail-az/state"
)

// A representation of an AWS service state that can be
// validated and stored with StateManager
type ConsistentStateService interface {
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

// Validates all required fields for service selector have been provided
func ValidateServiceSelector(selector ServiceSelector) error {

	if selector.Filter != "" && len(selector.Tags) > 0 {
		return fmt.Errorf("Validation failed: Both 'filter' and 'tags' selectors specified. Only one allowed.")
	}

	if selector.Filter == "" && len(selector.Tags) == 0 {
		return fmt.Errorf("Validation failed: One of 'filter' and 'tags' selectors must be specified.")
	}

	return nil
}
