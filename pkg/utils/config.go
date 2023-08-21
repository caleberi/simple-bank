package utils

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver  string `mapstructure:"DB_DRIVER"`
	DBSource  string `mapstructure:"DB_SOURCE"`
	VBankAddr string `mapstructure:"VBANK_ADDR"`
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
