package testutils

import (
	"ralts-cms/internal/machine"
	"ralts-cms/internal/maintenance"
	"time"
)

// CreateMachine creates a basic machine with default values
func CreateMachine(serialNumber string) *machine.Machine {
	return &machine.Machine{
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
		PpmStatus:       "Scheduled",
		TncDate:         time.Now().UTC().Format(time.RFC3339),
		PpmDate:         time.Now().AddDate(0, 1, 0).UTC().Format(time.RFC3339),
	}
}

// CreateMachineWithCustomFields creates a machine with custom field values
func CreateMachineWithCustomFields(serialNumber, customer, status string) *machine.Machine {
	machine := CreateMachine(serialNumber)
	machine.Customer = customer
	machine.Status = status
	return machine
}

// CreateMinimalMachine creates a machine with only required fields
func CreateMinimalMachine(serialNumber string) *machine.Machine {
	return &machine.Machine{
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
		WorkOrderDate:       time.Now().UTC().Format(time.RFC3339),
		ActionTaken:         "Routine maintenance",
		ReportedBy:          "John Doe",
		WorkerOrderType:     "Preventive",
		Attachment:          "maintenance.pdf",
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
