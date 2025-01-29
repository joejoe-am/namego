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

var configs *Configs

// LoadConfigs loads the configuration once
func LoadConfigs() {
	configPath := findConfigPath()
	if configPath == "" {
		log.Fatal("config.yaml not found in project root or any parent directory")
	}

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
		log.Fatal("missing required configuration fields in config.yaml")
	}
}

// GetConfigs returns the loaded configuration
func GetConfigs() *Configs {
	if configs == nil {
		LoadConfigs()
	}
	return configs
}

// findConfigPath searches for config.yaml in project root or parent directories
func findConfigPath() string {
	// Check if CONFIG_PATH is set
	if envPath, exists := os.LookupEnv("CONFIG_PATH"); exists {
		return envPath
	}

	// Start searching from the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	return searchConfigFile(workingDir)
}

// searchConfigFile recursively goes up directories to find config.yaml
func searchConfigFile(startDir string) string {
	for dir := startDir; dir != "/"; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return ""
}
