package maintenance

import (
	"context"
	"fmt"

	pkgdynamodb "ralts-cms/pkg/dynamodb"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	MaintenanceSortKeyPrefix = "Maintenance#"
)

//go:generate mockgen -destination=../maintenance/mock_maintenance_repository.go -package=maintenance -source=repository.go
type Repository interface {
	GetByWorkOrder(ctx context.Context, machineSerialNumber, workOrderNumber string) (*Maintenance, error)
	ListByMachine(ctx context.Context, machineSerialNumber string) ([]*Maintenance, error)
	Create(ctx context.Context, maintenance *Maintenance) error
	Update(ctx context.Context, maintenance *Maintenance) error
	Delete(ctx context.Context, machineSerialNumber, workOrderNumber string) error
}

type db struct {
	client *dynamodb.Client
	table  string
}

func NewRepository(client *dynamodb.Client, table string) Repository {
	return &db{
		client: client,
		table:  table,
	}
}

func (r *db) GetByWorkOrder(ctx context.Context, machineSerialNumber, workOrderNumber string) (*Maintenance, error) {
	maintenance := &Maintenance{
		MachineSerialNumber: machineSerialNumber,
		WorkOrderNumber:     workOrderNumber,
	}

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			pkgdynamodb.PartitionKey: &types.AttributeValueMemberS{Value: maintenance.GetPartitionKey()},
			pkgdynamodb.SortKey:      &types.AttributeValueMemberS{Value: maintenance.GetSortKey()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("maintenance not found")
	}

	var retrievedMaintenance Maintenance
	err = attributevalue.UnmarshalMap(result.Item, &retrievedMaintenance)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal maintenance: %w", err)
	}

	return &retrievedMaintenance, nil
}

func (r *db) ListByMachine(ctx context.Context, machineSerialNumber string) ([]*Maintenance, error) {
	maintenance := &Maintenance{MachineSerialNumber: machineSerialNumber}

	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :pk AND begins_with(%s, :sk)", pkgdynamodb.PartitionKey, pkgdynamodb.SortKey)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: maintenance.GetPartitionKey()},
			":sk": &types.AttributeValueMemberS{Value: MaintenanceSortKeyPrefix},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query maintenance records: %w", err)
	}

	var maintenanceRecords []*Maintenance
	for _, item := range result.Items {
		var maintenance Maintenance
		err := attributevalue.UnmarshalMap(item, &maintenance)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal maintenance: %w", err)
		}
		maintenanceRecords = append(maintenanceRecords, &maintenance)
	}

	return maintenanceRecords, nil
}

func (r *db) Create(ctx context.Context, maintenance *Maintenance) error {
	maintenance.SetTimestamps()

	item, err := attributevalue.MarshalMap(maintenance)
	if err != nil {
		return fmt.Errorf("failed to marshal maintenance: %w", err)
	}

	// Add DynamoDB keys
	item[pkgdynamodb.PartitionKey] = &types.AttributeValueMemberS{Value: maintenance.GetPartitionKey()}
	item[pkgdynamodb.SortKey] = &types.AttributeValueMemberS{Value: maintenance.GetSortKey()}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.table),
		Item:                item,
		ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s) AND attribute_not_exists(%s)", pkgdynamodb.PartitionKey, pkgdynamodb.SortKey)),
	})
	if err != nil {
		return fmt.Errorf("failed to create maintenance: %w", err)
	}

	return nil
}

func (r *db) Update(ctx context.Context, maintenance *Maintenance) error {
	maintenance.SetTimestamps()

	item, err := attributevalue.MarshalMap(maintenance)
	if err != nil {
		return fmt.Errorf("failed to marshal maintenance: %w", err)
	}

	// Add DynamoDB keys
	item[pkgdynamodb.PartitionKey] = &types.AttributeValueMemberS{Value: maintenance.GetPartitionKey()}
	item[pkgdynamodb.SortKey] = &types.AttributeValueMemberS{Value: maintenance.GetSortKey()}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update maintenance: %w", err)
	}

	return nil
}

func (r *db) Delete(ctx context.Context, machineSerialNumber, workOrderNumber string) error {
	maintenance := &Maintenance{
		MachineSerialNumber: machineSerialNumber,
		WorkOrderNumber:     workOrderNumber,
	}

	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			pkgdynamodb.PartitionKey: &types.AttributeValueMemberS{Value: maintenance.GetPartitionKey()},
			pkgdynamodb.SortKey:      &types.AttributeValueMemberS{Value: maintenance.GetSortKey()},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete maintenance: %w", err)
	}

	return nil
}
