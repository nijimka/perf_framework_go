package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TargetConfig struct {
	BaseURL   string `yaml:"base_url"`
	TimeoutMs int    `yaml:"timeout_ms"`
}

type LoadProfile struct {
	Type              string `yaml:"type"` // constant или ramp-up
	RampUpDurationSec int    `yaml:"ramp_up_duration_sec"`
	TargetWorkers     int    `yaml:"target_workers"`
}

type StepConfig struct {
	Name      string `yaml:"name"`
	Method    string `yaml:"method"`
	Path      string `yaml:"path"`
	SaveToken bool   `yaml:"save_token"`
	UseToken  bool   `yaml:"use_token"`
}

type ScenarioConfig struct {
	Name       string       `yaml:"name"`
	Iterations int          `yaml:"iterations"`
	Steps      []StepConfig `yaml:"steps"`
}

type Config struct {
	Target      TargetConfig     `yaml:"target"`
	LoadProfile LoadProfile      `yaml:"load_profile"`
	Scenarios   []ScenarioConfig `yaml:"scenarios"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if targetURL := os.Getenv("TARGET_URL"); targetURL != "" {
		cfg.Target.BaseURL = targetURL
	}

	return &cfg, nil
}
