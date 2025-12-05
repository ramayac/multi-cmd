package config

import (
	"fmt"
	"os"

	"github.com/ramayac/multi-cmd/internal/models"
	"gopkg.in/yaml.v3"
)

// Load reads and parses the configuration file
func Load(configPath string) (*models.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg models.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(cfg.Commands) == 0 {
		return nil, fmt.Errorf("no commands defined in config file")
	}

	return &cfg, nil
}
