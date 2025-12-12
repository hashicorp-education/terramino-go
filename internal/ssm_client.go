package terraminogo

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type RedisCredentials struct {
	IP       string `json:"redis_ip"`
	Port     string `json:"redis_port"`
	Password string `json:"redis_password"`
}

func GetSecret(secretName string) (RedisCredentials, error) {
	ctx := context.Background()
	var creds RedisCredentials

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return creds, err
	}

	client := secretsmanager.NewFromConfig(cfg)

	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return creds, err
	}

	err = json.Unmarshal([]byte(*out.SecretString), &creds)
	if err != nil {
		return creds, nil
	}

	return creds, nil

}
