package domain

import "github.com/mcastellin/aws-fail-az/state"

// A representation of an AWS service state that can be
// validated and stored with StateManager
type ConsistentServiceState interface {
	Check() (bool, error)
	Save(*state.StateManager) error
	Fail() error
	Restore([]byte) error
}
