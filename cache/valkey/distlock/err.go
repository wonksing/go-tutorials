package distlock

import "fmt"

type AcquireLockError struct {
	Msg string
}

func NewAcquireLockError(msg string) *AcquireLockError {
	return &AcquireLockError{Msg: msg}
}

func AsAcquireLockError(msg string, err error) *AcquireLockError {
	if err == nil {
		return &AcquireLockError{Msg: msg}
	}
	if v, ok := err.(*AcquireLockError); ok {
		v.Msg = fmt.Sprintf("%s: %s", msg, v.Msg)
		return v
	}

	return &AcquireLockError{Msg: fmt.Sprintf("%s: %s", msg, err.Error())}
}

func (e *AcquireLockError) Error() string {
	if e.Msg == "" {
		return "distlock acquire"
	}
	return fmt.Sprintf("distlock acquire: %s", e.Msg)
}
