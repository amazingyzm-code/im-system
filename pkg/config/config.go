package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Redis  RedisConfig  `mapstructure:"redis"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Kafka  KafkaConfig  `mapstructure:"kafka"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
}

type ServerConfig struct {
	Port   int    `mapstructure:"port"`
	Mode   string `mapstructure:"mode"`
	NodeID string `mapstructure:"node_id"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

var Global = &Config{}

func Init(cfgFile string) error {
	viper.SetConfigFile(cfgFile)
	// 环境变量优先级高于配置文件，KEY_SUB 对应 key.sub
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return viper.Unmarshal(Global)
}
