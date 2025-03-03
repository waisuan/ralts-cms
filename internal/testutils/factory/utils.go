package factory

import (
	"github.com/google/uuid"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/posts"
	"time"
)

func BuildMachine() *machines.Machine {
	today := time.Date(2024, time.January, 11, 0, 0, 0, 0, time.UTC)

	return &machines.Machine{
		SerialNumber:    TestUUID(),
		Customer:        "Evan",
		State:           "Test",
		AccountType:     "Custom",
		Model:           "A",
		Status:          "CONFIRMED",
		Brand:           "ASUS",
		District:        "Test",
		PersonInCharge:  "Naus",
		ReportedBy:      "Naus",
		AdditionalNotes: "Lorem Ipsum is simply dummy text of the printing and typesetting industry",
		Attachment:      "test-file.pdf",
		PpmStatus:       "PENDING",
		TncDate:         &today,
		PpmDate:         &today,
	}
}

func BuildPost() *posts.Post {
	return &posts.Post{
		Title:   "Test",
		Content: "Lorem Ipsum is simply dummy text of the printing and typesetting industry",
		UserID:  1,
	}
}

func TestUUID() string {
	return uuid.New().String()
}
