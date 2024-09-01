package machines

import (
	"time"
)

type Machine struct {
	ID              uint       `json:"id"`
	SerialNumber    string     `gorm:"uniqueIndex;unique;not null;default:null" json:"serial_number"`
	Customer        string     `json:"customer"`
	State           string     `json:"state"`
	AccountType     string     `json:"account_type"`
	Model           string     `json:"model"`
	Status          string     `json:"status"`
	Brand           string     `json:"brand"`
	District        string     `json:"district"`
	PersonInCharge  string     `json:"person_in_charge"`
	ReportedBy      string     `json:"reported_by"`
	AdditionalNotes string     `json:"additional_notes"`
	Attachment      string     `json:"attachment"`
	PpmStatus       string     `json:"ppm_status"`
	TncDate         *time.Time `json:"tnc_date"`
	PpmDate         *time.Time `json:"ppm_date"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (m *Machine) FormattedTncDate() string {
	if m.TncDate == nil {
		return ""
	}
	return m.TncDate.Format("2006-01-02")
}

func (m *Machine) FormattedPpmDate() string {
	if m.PpmDate == nil {
		return ""
	}
	return m.PpmDate.Format("2006-01-02")
}
