package config

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var validate = validator.New()

type PetConfig struct {
	Name      string `yaml:"name" validate:"required"`
	Type      string `yaml:"type" validate:"required,oneof=pishock lovense"`
	ShareCode string `yaml:"share_code,omitempty"` // For PiShock
	LovenseID string `yaml:"lovense_id,omitempty"` // For Lovense
	LovenseIP string `yaml:"lovense_ip,omitempty"` // For Lovense (local IP)
}

type Config struct {
	LogLevel        string      `yaml:"log_level" validate:"required,oneof=debug info warn error"`
	Theme           string      `yaml:"theme" validate:"required,oneof=base base16 catppuccin charm dracula"`
	PiShockAPIKey   string      `yaml:"pishock_api_key"`
	PiShockAppName  string      `yaml:"pishock_app_name"`
	ShockerID		string		`yaml:"pishock_shocker_id"`
	Pets            []PetConfig `yaml:"pets" validate:"dive"`
}

func DefaultConfig() *Config {
	return &Config{
		LogLevel:        "info",
		Theme:           "dracula",
		PiShockAPIKey:	"your-api-key-here",
		PiShockAppName:	"GolangPetController",
		ShockerID:		"shocker-id-here",
		Pets: []PetConfig{
			{
				Name:      "DefaultShock",
				Type:      "pishock",
			},
		},
	}
}

type ConfigManager struct {
	config *Config
	mu     sync.RWMutex
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: &Config{},
	}
}

func (m *ConfigManager) Load(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			m.config = DefaultConfig()
			if valErr := validate.Struct(m.config); valErr != nil {
				return fmt.Errorf("default config validation failed: %w", valErr)
			}
			return m.saveUnlocked(path)
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	newConfig := &Config{}
	if err := yaml.Unmarshal(data, newConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validate.Struct(newConfig); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	m.config = newConfig
	return nil
}

func (m *ConfigManager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.config
}

func (m *ConfigManager) Update(fn func(c *Config)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return errors.New("configuration not loaded")
	}

	updatedConfig := *m.config
	fn(&updatedConfig)

	if err := validate.Struct(&updatedConfig); err != nil {
		return fmt.Errorf("updated config validation failed: %w", err)
	}

	m.config = &updatedConfig
	return nil
}

func (m *ConfigManager) Save(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.saveUnlocked(path)
}

func (m *ConfigManager) saveUnlocked(path string) error {
	if m.config == nil {
		return errors.New("configuration not loaded")
	}

	if err := validate.Struct(m.config); err != nil {
		return fmt.Errorf("config validation failed before saving: %w", err)
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
