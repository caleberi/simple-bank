package utils

import (
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	VBankAddr            string        `mapstructure:"VBANK_ADDR"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

var cfg = &Config{}
var once sync.Once

func LoadConfig(path, filename, envType string) *Config {
	once.Do(func() {
		viper.AddConfigPath(path)
		viper.SetConfigName(filename)
		viper.SetConfigType(envType)

		err := viper.ReadInConfig()
		if err != nil {
			log.Fatal(err)
		}

		err = viper.Unmarshal(cfg)

		if err != nil {
			log.Fatal(err)
		}
	})
	return cfg
}
