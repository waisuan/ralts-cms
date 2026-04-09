#!/bin/bash

# Wait for LocalStack to be ready
echo "Waiting for LocalStack to be ready..."
until curl -f http://localhost:4566/_localstack/health > /dev/null 2>&1; do
    echo "LocalStack not ready yet, waiting..."
    sleep 2
done

echo "LocalStack is ready. Creating S3 bucket..."

# Create the S3 bucket for attachments
awslocal s3api create-bucket \
    --bucket ralts-cms-attachments \
    --region us-east-1

# Verify bucket was created
if awslocal s3api head-bucket --bucket ralts-cms-attachments 2>/dev/null; then
    echo "✅ S3 bucket 'ralts-cms-attachments' created successfully"
else
    echo "❌ Failed to create S3 bucket"
    exit 1
fi

# Optional: Set up CORS configuration for the bucket
echo "Setting up CORS configuration..."
cat > /tmp/cors-config.json << EOF
{
  "CORSRules": [
    {
      "AllowedHeaders": ["*"],
      "AllowedMethods": ["GET", "POST", "PUT", "DELETE", "HEAD"],
      "AllowedOrigins": [
        "http://localhost:3000",
        "http://localhost:8080",
        "https://app.localstack.cloud",
        "http://app.localstack.cloud"
      ],
      "ExposeHeaders": ["ETag", "x-amz-server-side-encryption", "x-amz-request-id", "x-amz-id-2"],
      "MaxAgeSeconds": 86400
    }
  ]
}
EOF

awslocal s3api put-bucket-cors \
    --bucket ralts-cms-attachments \
    --cors-configuration file:///tmp/cors-config.json

echo "✅ CORS configuration applied to bucket"

# List buckets to confirm
echo "Available S3 buckets:"
awslocal s3api list-buckets --query 'Buckets[].Name' --output table

echo "LocalStack S3 setup completed successfully!"
