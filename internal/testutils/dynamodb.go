package testutils

import (
	"context"

	pkgdynamodb "ralts-cms/pkg/dynamodb"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func ClearTable(ctx context.Context, tableName string, client *dynamodb.Client) error {
	// Scan the table to get all items
	scanOutput, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return err
	}

	// Delete all items
	for _, item := range scanOutput.Items {
		pkAttr, ok := item[pkgdynamodb.PartitionKey].(*types.AttributeValueMemberS)
		if !ok {
			continue
		}
		skAttr, ok := item[pkgdynamodb.SortKey].(*types.AttributeValueMemberS)
		if !ok {
			continue
		}

		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				pkgdynamodb.PartitionKey: &types.AttributeValueMemberS{Value: pkAttr.Value},
				pkgdynamodb.SortKey:      &types.AttributeValueMemberS{Value: skAttr.Value},
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
