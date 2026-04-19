package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 瀹氫箟鎵€鏈夊彲閰嶇疆椤?
type Config struct {
	Port          int    `yaml:"port"`
	Ip            string `yaml:"ip"`
	WSPort        int    `yaml:"wsport"`
	DatabaseURL   string `yaml:"database_url"`
	JWTKey        string `yaml:"jwt_key"`
	JWTAlgorithm  string `yaml:"jwt_algorithm"`
	JWTExpiry     string `yaml:"jwt_expiry"`
	AdminPassword string `yaml:"admin_password"`
}

var cfg *Config
var FirstRun bool

var (
	Host          string
	Port          int
	WSPort        int
	DatabaseURL   string
	JWTKey        []byte
	JWTAlgorithm  string
	JWTExpiry     time.Duration
	AdminPassword string
)

const (
	defaultPort          = 5231
	defaultIp            = "127.0.0.1"
	defaultWSPort        = 5232
	defaultDatabaseURL   = "root:1234567890@tcp(127.0.0.1:3306)/users?charset=utf8mb4&parseTime=True&loc=Local"
	defaultJWTKey        = "CometKards-is-a-help-much-kards-players-that-can't-find-gameuser-or-baned"
	defaultJWTAlgorithm  = "HS256"
	defaultJWTExpiry     = "24h"
	defaultAdminPassword = "change-this-password"
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
	AdminPassword = cfg.AdminPassword
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
	hadAdminPasswordInFile := cfgFromFile.AdminPassword != ""

	applyEnvOverrides(cfgFromFile)

	applyDefaults(cfgFromFile)

	// 濡傛灉閰嶇疆鏂囦欢涓嶅瓨鍦紝鍒欏垱寤哄畠
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		FirstRun = true
		if err := saveYAMLConfig(cfgFromFile); err != nil {
			return nil, fmt.Errorf("failed to create config.yaml: %w", err)
		}
	} else if !hadAdminPasswordInFile && os.Getenv("ADMIN_PASSWORD") == "" {
		if err := saveYAMLConfig(cfgFromFile); err != nil {
			return nil, fmt.Errorf("failed to update config.yaml with admin_password: %w", err)
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
	if v := getEnv("ADMIN_PASSWORD", ""); v != "" {
		cfg.AdminPassword = v
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
	if cfg.AdminPassword == "" {
		cfg.AdminPassword = defaultAdminPassword
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func PromptInitialSetup() error {
	if !FirstRun && cfg.DatabaseURL != "" && cfg.Ip != "" && cfg.Port != 0 {
		return nil
	}

	fmt.Println("First-time setup detected. Please configure server parameters.")
	fmt.Print("MySQL URL (user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local): ")
	var dbURL string
	fmt.Scanln(&dbURL)
	if dbURL == "" {
		dbURL = defaultDatabaseURL
	}

	fmt.Print("Server IP (default 127.0.0.1): ")
	var ip string
	fmt.Scanln(&ip)
	if ip == "" {
		ip = defaultIp
	}

	fmt.Print("Server port (default 5231): ")
	var port int
	_, err := fmt.Scanln(&port)
	if err != nil || port == 0 {
		port = defaultPort
	}

	fmt.Print("WebSocket port (default 5232): ")
	var wsPort int
	_, err = fmt.Scanln(&wsPort)
	if err != nil || wsPort == 0 {
		wsPort = defaultWSPort
	}

	cfg.DatabaseURL = dbURL
	cfg.Ip = ip
	cfg.Port = port
	cfg.WSPort = wsPort
	if cfg.AdminPassword == "" {
		cfg.AdminPassword = defaultAdminPassword
	}

	if err := saveYAMLConfig(cfg); err != nil {
		return err
	}

	Host = cfg.Ip
	Port = cfg.Port
	WSPort = cfg.WSPort
	DatabaseURL = cfg.DatabaseURL
	AdminPassword = cfg.AdminPassword

	fmt.Println("Configuration saved. Please restart the server.")
	return nil
}

func GetKardsTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}
