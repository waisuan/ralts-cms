package machine

import (
	"context"
	"fmt"

	pkgdynamodb "ralts-cms/pkg/dynamodb"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

//go:generate mockgen -destination=../machine/mock_machines_repository.go -package=machine -source=repository.go
type Repository interface {
	GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error)
	List(ctx context.Context, limit int32) ([]*Machine, error)
	Create(ctx context.Context, machine *Machine) error
	Update(ctx context.Context, machine *Machine) error
	Delete(ctx context.Context, serialNumber string) error
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

func (r *db) GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error) {
	machine := &Machine{SerialNumber: serialNumber}

	item, err := attributevalue.MarshalMap(machine)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal machine: %w", err)
	}

	// Add DynamoDB keys
	item[pkgdynamodb.PartitionKey] = &types.AttributeValueMemberS{Value: machine.GetPartitionKey()}
	item[pkgdynamodb.SortKey] = &types.AttributeValueMemberS{Value: machine.GetSortKey()}

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			pkgdynamodb.PartitionKey: &types.AttributeValueMemberS{Value: machine.GetPartitionKey()},
			pkgdynamodb.SortKey:      &types.AttributeValueMemberS{Value: machine.GetSortKey()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("machine not found")
	}

	var retrievedMachine Machine
	err = attributevalue.UnmarshalMap(result.Item, &retrievedMachine)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal machine: %w", err)
	}

	return &retrievedMachine, nil
}

func (r *db) List(ctx context.Context, limit int32) ([]*Machine, error) {
	// Build scan input
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(r.table),
		Limit:     aws.Int32(limit),
	}

	// Add filter expression to only get machine records (not maintenance records)
	scanInput.FilterExpression = aws.String("begins_with(#pk, :pkPrefix)")
	scanInput.ExpressionAttributeNames = map[string]string{
		"#pk": pkgdynamodb.PartitionKey,
	}
	scanInput.ExpressionAttributeValues = map[string]types.AttributeValue{
		":pkPrefix": &types.AttributeValueMemberS{Value: "Machine#"},
	}

	result, err := r.client.Scan(ctx, scanInput)
	if err != nil {
		return nil, fmt.Errorf("failed to scan machines: %w", err)
	}

	var machines []*Machine
	for _, item := range result.Items {
		var machine Machine
		err := attributevalue.UnmarshalMap(item, &machine)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal machine: %w", err)
		}
		machines = append(machines, &machine)
	}

	return machines, nil
}

func (r *db) Create(ctx context.Context, machine *Machine) error {
	machine.SetTimestamps()

	item, err := attributevalue.MarshalMap(machine)
	if err != nil {
		return fmt.Errorf("failed to marshal machine: %w", err)
	}

	// Add DynamoDB keys
	item[pkgdynamodb.PartitionKey] = &types.AttributeValueMemberS{Value: machine.GetPartitionKey()}
	item[pkgdynamodb.SortKey] = &types.AttributeValueMemberS{Value: machine.GetSortKey()}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.table),
		Item:                item,
		ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s) AND attribute_not_exists(%s)", pkgdynamodb.PartitionKey, pkgdynamodb.SortKey)),
	})
	if err != nil {
		return fmt.Errorf("failed to create machine: %w", err)
	}

	return nil
}

func (r *db) Update(ctx context.Context, machine *Machine) error {
	machine.SetTimestamps()

	item, err := attributevalue.MarshalMap(machine)
	if err != nil {
		return fmt.Errorf("failed to marshal machine: %w", err)
	}

	// Add DynamoDB keys
	item[pkgdynamodb.PartitionKey] = &types.AttributeValueMemberS{Value: machine.GetPartitionKey()}
	item[pkgdynamodb.SortKey] = &types.AttributeValueMemberS{Value: machine.GetSortKey()}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update machine: %w", err)
	}

	return nil
}

func (r *db) Delete(ctx context.Context, serialNumber string) error {
	machine := &Machine{SerialNumber: serialNumber}

	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			pkgdynamodb.PartitionKey: &types.AttributeValueMemberS{Value: machine.GetPartitionKey()},
			pkgdynamodb.SortKey:      &types.AttributeValueMemberS{Value: machine.GetSortKey()},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}

	return nil
}
