package errs

import "errors"

var ErrCritical = errors.New("critical error, signaling shutdown")
