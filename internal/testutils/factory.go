// Package testutils provides factory functions for creating test data
// including machines, maintenance records, and users for testing purposes.
package testutils

import (
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"ralts-cms/internal/users"
	"time"
)

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}

// CreateMachine creates a basic machine with default values
func CreateMachine(serialNumber string) *machines.Machine {
	return &machines.Machine{
		SerialNumber:    serialNumber,
		Customer:        "Test Customer",
		State:           "Active",
		AccountType:     "Standard",
		Model:           "Test Model",
		Status:          "Operational",
		Brand:           "Test Brand",
		District:        "Test District",
		PersonInCharge:  "John Doe",
		ReportedBy:      "Jane Smith",
		AdditionalNotes: "Test notes",
		Attachment:      "test.pdf",
		TncDate:         time.Now().UTC(),
		PpmDate:         time.Now().AddDate(0, 1, 0).UTC(),
	}
}

// CreateMachineWithCustomFields creates a machine with custom field values
func CreateMachineWithCustomFields(serialNumber, customer, status string) *machines.Machine {
	machine := CreateMachine(serialNumber)
	machine.Customer = customer
	machine.Status = status
	return machine
}

// CreateMinimalMachine creates a machine with only required fields
func CreateMinimalMachine(serialNumber string) *machines.Machine {
	return &machines.Machine{
		SerialNumber: serialNumber,
		Customer:     "Minimal Customer",
		Status:       "Active",
	}
}

// CreateMaintenance creates a basic maintenance record with default values
func CreateMaintenance(machineSerialNumber, workOrderNumber string) *maintenance.Maintenance {
	return &maintenance.Maintenance{
		MachineSerialNumber: machineSerialNumber,
		WorkOrderNumber:     workOrderNumber,
		WorkOrderDate:       time.Now().UTC(),
		ActionTaken:         "Routine maintenance",
		ReportedBy:          "John Doe",
		WorkOrderType:       "Preventive",
		Attachment:          StringPtr("maintenance.pdf"),
	}
}

// CreateMaintenanceWithCustomFields creates a maintenance record with custom field values
func CreateMaintenanceWithCustomFields(machineSerialNumber, workOrderNumber, actionTaken, reportedBy string) *maintenance.Maintenance {
	maintenance := CreateMaintenance(machineSerialNumber, workOrderNumber)
	maintenance.ActionTaken = actionTaken
	maintenance.ReportedBy = reportedBy
	return maintenance
}

// CreateMinimalMaintenance creates a maintenance record with only required fields
func CreateMinimalMaintenance(machineSerialNumber, workOrderNumber string) *maintenance.Maintenance {
	return &maintenance.Maintenance{
		MachineSerialNumber: machineSerialNumber,
		WorkOrderNumber:     workOrderNumber,
		ActionTaken:         "Minimal action",
		ReportedBy:          "Minimal Tech",
	}
}

// CreateUser creates a basic user with default values
func CreateUser(username, password string) *users.User {
	return &users.User{
		Username: username,
		Email:    username + "@example.com", // Generate email from username
		Password: password,
		Role:     "", // Will be set to default by repository
		Approved: false,
		Status:   nil, // Will be set to default by repository
		Avatar:   nil,
	}
}

// CreateUserWithCustomFields creates a user with custom field values
func CreateUserWithCustomFields(username, password, role, status string) *users.User {
	user := CreateUser(username, password)
	user.Role = role
	user.Status = &status
	return user
}

// CreateUserWithAvatar creates a user with avatar
func CreateUserWithAvatar(username, password string, avatar *string) *users.User {
	user := CreateUser(username, password)
	user.Avatar = avatar
	return user
}

// CreateMinimalUser creates a user with only required fields
func CreateMinimalUser(username, password string) *users.User {
	return &users.User{
		Username: username,
		Email:    username + "@example.com",
		Password: password,
	}
}
