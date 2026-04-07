package machines

import "errors"

// ErrNotFound is returned when a machine does not exist for the requested identifier.
var ErrNotFound = errors.New("machine not found")
