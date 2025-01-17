package errorz

import "errors"

var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrNeedRetry        = errors.New("need retry")
)
