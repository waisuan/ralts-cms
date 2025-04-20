package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"ralts-cms/internal/deps"
)

func main() {
	cfg, err := deps.LoadConfig()
	if err != nil {
		panic(err)
	}

	db, err := deps.InitDynamoDBClient(cfg)
	if err != nil {
		panic(err)
	}

	o, err := db.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String("ralts.dev"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(o.Table.TableName)
}
