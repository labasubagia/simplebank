package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment              string        `mapstructure:"ENVIRONMENT"`
	DBDriver                 string        `mapstructure:"DB_DRIVER"`
	DBSource                 string        `mapstructure:"DB_SOURCE"`
	DBMigrationURL           string        `mapstructure:"DB_MIGRATION_URL"`
	RedisAddress             string        `mapstructure:"REDIS_ADDRESS"`
	HTTPServerAddress        string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	HTTPGatewayServerAddress string        `mapstructure:"HTTP_GATEWAY_SERVER_ADDRESS"`
	GRPCServerAddress        string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey        string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration      time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration     time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func (c Config) IsEnvProduction() bool {
	return c.Environment == "production"
}

func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
