package graph

import "ralts-cms/internal/machines"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	MachineResolver *machines.Resolver
}
