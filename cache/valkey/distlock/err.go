package distlock

import "fmt"

type AcquireLockError struct {
	Msg string
}

func NewAcquireLockError(msg string) *AcquireLockError {
	return &AcquireLockError{Msg: msg}
}

func (e *AcquireLockError) Error() string {
	if e.Msg == "" {
		return "acquire lock"
	}
	return fmt.Sprintf("acquire lock: %s", e.Msg)
}
