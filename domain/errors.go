package domain

// An error type to signal the activity has failed and whether or not
// it can be retried
type ActivityFailedError struct {
	Wrap      error
	Temporary bool
}

func (e ActivityFailedError) Error() string {
	return e.Wrap.Error()
}

func (e ActivityFailedError) IsTemporary() bool {
	return e.Temporary
}

// An error type to signal the current activity has failed and that the
// program execution should be interrupted as soon as possible
type InterruptExecutionError struct {
	Wrap error
}

func (e InterruptExecutionError) Error() string {
	return e.Wrap.Error()
}
