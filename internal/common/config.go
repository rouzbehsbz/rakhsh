package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	MachineId uint16 `mapstructure:"MACHINE_ID"`
	Host      string `mapstructure:"HOST"`
	Port      uint16 `mapstructure:"PORT"`

	PostgresShard1         string `mapstructure:"POSTGRES_SHARD_1"`
	PostgresShard2         string `mapstructure:"POSTGRES_SHARD_2"`
	PostgresShard3         string `mapstructure:"POSTGRES_SHARD_3"`
	PostgresMaxConnections int    `mapstructure:"POSTGRES_MAX_CONNECTIONS"`
}

func NewConfig(isDevMode bool) (*Config, error) {
	if isDevMode {
		return newConfigFromEnvFile("./.env")
	} else {
		return newConfigFromOsEnv()
	}
}

func newConfigFromEnvFile(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("env")
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config

	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func newConfigFromOsEnv() (*Config, error) {
	v := viper.New()

	v.AutomaticEnv()

	envs := os.Environ()

	for _, env := range envs {
		pair := strings.SplitN(env, "=", 2)

		if len(pair) != 2 {
			println(len(pair))
			return nil, fmt.Errorf("invalid env pair found in os: %v", pair)
		}

		v.Set(pair[0], pair[1])
	}

	var c Config

	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
