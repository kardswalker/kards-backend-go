package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 定义所有可配置项
type Config struct {
	Port         int    `yaml:"port"`
	Ip           string `yaml:"ip"`
	WSPort       int    `yaml:"wsport"`
	DatabaseURL  string `yaml:"database_url"`
	JWTKey       string `yaml:"jwt_key"`
	JWTAlgorithm string `yaml:"jwt_algorithm"`
	JWTExpiry    string `yaml:"jwt_expiry"`
}

var cfg *Config

var (
	Host         string
	Port         int
	WSPort       int
	DatabaseURL  string
	JWTKey       []byte
	JWTAlgorithm string
	JWTExpiry    time.Duration
)

const (
	defaultPort         = 5231
	defaultIp           = "127.0.0.1"
	defaultWSPort       = 5232
	defaultDatabaseURL  = "root:1234567890@tcp(127.0.0.1:3306)/users?charset=utf8mb4&parseTime=True&loc=Local"
	defaultJWTKey       = "CometKards-is-a-help-much-kards-players-that-can't-find-gameuser-or-baned"
	defaultJWTAlgorithm = "HS256"
	defaultJWTExpiry    = "24h"
)

func init() {
	var err error
	cfg, err = LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	Host = cfg.Ip
	Port = cfg.Port
	WSPort = cfg.WSPort
	DatabaseURL = cfg.DatabaseURL
	JWTKey = []byte(cfg.JWTKey)
	JWTAlgorithm = cfg.JWTAlgorithm
	var dur time.Duration
	if dur, err = time.ParseDuration(cfg.JWTExpiry); err != nil {
		panic("invalid jwt_expiry: " + err.Error())
	}
	JWTExpiry = dur
}

func LoadConfig() (*Config, error) {
	cfgFromFile := &Config{}
	if err := loadYAMLConfig(cfgFromFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	applyEnvOverrides(cfgFromFile)

	applyDefaults(cfgFromFile)

	// 如果配置文件不存在，则创建它
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := saveYAMLConfig(cfgFromFile); err != nil {
			return nil, fmt.Errorf("failed to create config.yaml: %w", err)
		}
	}

	return cfgFromFile, nil
}

func loadYAMLConfig(cfg *Config) error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}
	return nil
}

func saveYAMLConfig(cfg *Config) error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func applyEnvOverrides(cfg *Config) {
	if v := getEnv("PORT", ""); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Port = i
		}
	}
	if v := getEnv("IP", ""); v != "" {
		cfg.Ip = v
	}
	if v := getEnv("WSPORT", ""); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.WSPort = i
		}
	}
	if v := getEnv("DB_URL", ""); v != "" {
		cfg.DatabaseURL = v
	}
	if v := getEnv("JWT_SECRET", ""); v != "" {
		cfg.JWTKey = v
	}
	if v := getEnv("JWT_ALGORITHM", ""); v != "" {
		cfg.JWTAlgorithm = v
	}
	if v := getEnv("JWT_EXPIRY", ""); v != "" {
		cfg.JWTExpiry = v
	}
}

func applyDefaults(cfg *Config) {
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	if cfg.Ip == "" {
		cfg.Ip = defaultIp
	}
	if cfg.WSPort == 0 {
		cfg.WSPort = defaultWSPort
	}
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = defaultDatabaseURL
	}
	if cfg.JWTKey == "" {
		cfg.JWTKey = defaultJWTKey
	}
	if cfg.JWTAlgorithm == "" {
		cfg.JWTAlgorithm = defaultJWTAlgorithm
	}
	if cfg.JWTExpiry == "" {
		cfg.JWTExpiry = defaultJWTExpiry
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetKardsTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}
