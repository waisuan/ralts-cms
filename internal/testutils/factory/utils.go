package factory

import (
	"github.com/google/uuid"
	"ralts-cms/internal/machines"
)

func BuildMachine() *machines.Machine {
	return &machines.Machine{
		SerialNumber: TestUUID(),
		Customer:     "Evan",
	}
}

func TestUUID() string {
	return uuid.New().String()
}
