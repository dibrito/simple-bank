package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Env               string        `mapstructure:"ENVIROMENT"`
	DBDriver          string        `mapstructure:"DB_DRIVER"`
	DBMigrationPath   string        `mapstructure:"MIGRATION"`
	DBSource          string        `mapstructure:"DB_SOURCE"`
	HttpAddress       string        `mapstructure:"HTTP_ADDRESS"`
	GRPCAddress       string        `mapstructure:"GRPC_ADDRESS"`
	TokenSymmetricKey string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessDuration    time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshDuration   time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

// LoadConfig read configuration from a file or enviromental variables.
func LoadConfig(path string) (c Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&c)
	return
}
