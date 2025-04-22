package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env         string        `yaml:"env" required:"true"`
	RestAddress int           `yaml:"rest_address" required:"true"`
	GrpcAddress int           `yaml:"grpc_address" required:"true"`
	StoragePath string        `yaml:"storage_path" required:"true"`
	Timeout     time.Duration `yaml:"timeout" required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" required:"true"`
	TimePeriod  time.Duration `yaml:"time_period" default:"1h"`
}

func MustLoad(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("read config file %s error: %v", configPath, err)
	}
	return &cfg
}
