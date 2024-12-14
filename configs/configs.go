package configs

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

// Configs holds application configuration
type Configs struct {
	ServiceName  string `yaml:"service_name"`
	RabbitMQURL  string `yaml:"rabbitmq_url"`
	ExchangeName string `yaml:"exchange_name"`
}

const (
	DefaultConfigFile = "configs/config.yaml"
)

var (
	configs *Configs
)

// LoadConfigs loads the configuration once
func LoadConfigs() {
	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	configPath := filepath.Join(workingDir, DefaultConfigFile)

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("failed to open config file %s: %v", configPath, err)
	}
	defer file.Close()

	configs = &Configs{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(configs); err != nil {
		log.Fatalf("failed to decode config file: %v", err)
	}

	if configs.ExchangeName == "" {
		configs.ExchangeName = "nameko-rpc"
	}

	// Validate required fields
	if configs.ServiceName == "" || configs.RabbitMQURL == "" {
		log.Fatal("missing required configuration fields in config.yml")
	}
}

// GetConfigs returns the loaded configuration
func GetConfigs() *Configs {
	if configs == nil {
		LoadConfigs()
	}
	return configs
}
