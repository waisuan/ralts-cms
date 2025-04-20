package deps

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func InitDynamoDBClient(cfg *Config) (*dynamodb.Client, error) {
	var (
		awsCfg aws.Config
		err    error
	)

	if cfg.Env == "development" || cfg.Env == "test" {
		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion("us-west-1"),
			config.WithBaseEndpoint(fmt.Sprintf("http://%s:%d", cfg.DB.Hostname, cfg.DB.Port)),
		)
	} else {
		// Load default AWS configuration
		awsCfg, err = config.LoadDefaultConfig(context.TODO())
	}

	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(awsCfg), nil
}
