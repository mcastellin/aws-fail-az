package domain

// A representation of an AWS service state that can be
// validated and stored with StateManager
type ConsistentServiceState interface {
	Check() (bool, error)
	Save() error
	Fail() error
	Restore() error
}
