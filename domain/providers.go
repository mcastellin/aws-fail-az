package domain

import "github.com/aws/aws-sdk-go-v2/aws"

type AWSProvider struct {
	awsConfig *aws.Config
}

func (provider AWSProvider) GetConnection() aws.Config {
	return *provider.awsConfig
}

// Creates a new provider from AWS configuration
func NewProviderFromConfig(cfg *aws.Config) AWSProvider {
	return AWSProvider{
		awsConfig: cfg,
	}
}
