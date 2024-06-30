package maintenance

import (
	"ralts-cms/internal/machines"
	"time"
)

type Maintenance struct {
	ID              uint
	MachineID       uint
	Machine         machines.Machine
	WorkOrderNumber string
	WorkOrderDate   *time.Time
	ActionTaken     string
	ReportedBy      string
	WorkerOrderType string
	Attachment      string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
