package maintenance

import "errors"

// ErrNotFound is returned when a maintenance record does not exist for the requested keys.
var ErrNotFound = errors.New("maintenance not found")
