package util

import "github.com/spf13/viper"

type Config struct {
	RedisEndpoint string `mapstructure:"REDIS_ENDPOINT"`
	SymbolCount   int    `mapstructure:"SYMBOL_COUNT"`
	Multiplyer    int    `mapstructure:"MULTIPLYER"`
	PayloadSize   int    `mapstructure:"PAYLOADSIZE"`
	Workers       int    `mapstructure:"WORKERS"`
}

// LoadConfig loads app.env if it exists and sets envars
func LoadConfig(path string) (config Config, err error) {
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
