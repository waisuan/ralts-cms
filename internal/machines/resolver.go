package machines

import (
	"context"
	"ralts-cms/graph/model"
	"time"
)

const (
	LabelSerialNumber   = "Serial No."
	LabelCustomer       = "Customer"
	LabelState          = "State"
	LabelAccountType    = "Account Type"
	LabelModel          = "Model"
	LabelStatus         = "Status"
	LabelBrand          = "Brand"
	LabelDistrict       = "District"
	LabelPersonInCharge = "Person In Charge"
	LabelReportedBy     = "Reported By"
	LabelPpmStatus      = "PPM Status"
	LabelTncDate        = "TNC Date"
	LabelPpmDate        = "PPM Date"
)

type Resolver struct {
	repo Repository
}

func NewResolver(repo Repository) *Resolver {
	return &Resolver{repo: repo}
}

func (r *Resolver) GetFormFieldsBySerialNumber(serialNumber string) ([]*model.FormField, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	machine, err := r.repo.GetBySerialNumber(ctx, serialNumber)
	if err != nil {
		return nil, ErrMachineNotFound
	}

	// TODO: Implement the rest of the fields
	// 	- AdditionalNotes, Attachment, CreatedAt, UpdatedAt
	fields := []*model.FormField{
		{
			Type:  model.FormFieldTypeText,
			Label: LabelSerialNumber,
			Value: machine.SerialNumber,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelCustomer,
			Value: machine.Customer,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelState,
			Value: machine.State,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelAccountType,
			Value: machine.AccountType,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelModel,
			Value: machine.Model,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelStatus,
			Value: machine.Status,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelBrand,
			Value: machine.Brand,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelDistrict,
			Value: machine.District,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelPersonInCharge,
			Value: machine.PersonInCharge,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelReportedBy,
			Value: machine.ReportedBy,
		},
		{
			Type:  model.FormFieldTypeText,
			Label: LabelPpmStatus,
			Value: machine.PpmStatus,
		},
		{
			Type:  model.FormFieldTypeCalendar,
			Label: LabelTncDate,
			Value: machine.FormattedTncDate(),
		},
		{
			Type:  model.FormFieldTypeCalendar,
			Label: LabelPpmDate,
			Value: machine.FormattedPpmDate(),
		},
	}

	return fields, nil
}
