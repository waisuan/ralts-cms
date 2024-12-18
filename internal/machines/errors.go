package machines

import (
	"errors"
)

var (
	// ErrMachineNotFound is returned when a machine is not found
	ErrMachineNotFound = errors.New("machine not found")
)
