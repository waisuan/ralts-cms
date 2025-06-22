#!/bin/bash

echo "Starting local DynamoDB instance..."
docker compose up --wait -d dynamodb-local

echo "Waiting for DynamoDB to be ready..."
sleep 5

echo "DynamoDB is running on http://localhost:8000"
echo "You can access the DynamoDB shell with: aws dynamodb --endpoint-url http://localhost:8000 --region us-east-1" 