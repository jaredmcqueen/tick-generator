package util

import (
	"log"

	"github.com/spf13/viper"
)

type Envars struct {
	RedisEndpoint string `mapstructure:"REDIS_ENDPOINT"`
	SymbolCount   int    `mapstructure:"SYMBOL_COUNT"`
}

var Config Envars

func init() {
	config, err := loadConfig(".")
	if err != nil {
		log.Fatal("cannot load configuration", err)
	}
	Config = config
}

// LoadConfig loads app.env if it exists and sets envars
func loadConfig(path string) (config Envars, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
