#!/bin/bash

echo "Creating DynamoDB table 'ralts'..."

aws dynamodb create-table \
    --table-name ralts \
    --attribute-definitions \
        AttributeName=PK,AttributeType=S \
        AttributeName=SK,AttributeType=S \
        AttributeName=GSI1_PK,AttributeType=S \
        AttributeName=GSI1_SK,AttributeType=S \
    --key-schema \
        AttributeName=PK,KeyType=HASH \
        AttributeName=SK,KeyType=RANGE \
    --global-secondary-indexes \
        '[{"IndexName":"gsi_1","KeySchema":[{"AttributeName":"GSI1_PK","KeyType":"HASH"},{"AttributeName":"GSI1_SK","KeyType":"RANGE"}],"Projection":{"ProjectionType":"ALL"}}]' \
    --billing-mode PAY_PER_REQUEST \
    --endpoint-url http://localhost:8000 \
    --region us-east-1

echo "Table creation completed!" 