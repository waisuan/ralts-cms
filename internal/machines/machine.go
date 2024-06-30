package machines

import (
	"time"
)

type Machine struct {
	ID              uint
	SerialNumber    string `gorm:"uniqueIndex;unique;not null;default:null"`
	Customer        string
	State           string
	AccountType     string
	Model           string
	Status          string
	Brand           string
	District        string
	PersonInCharge  string
	ReportedBy      string
	AdditionalNotes string
	Attachment      string
	PpmStatus       string
	TncDate         *time.Time
	PpmDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
