package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 瀹氫箟鎵€鏈夊彲閰嶇疆椤?
type Config struct {
	Port            int      `yaml:"port"`
	Ip              string   `yaml:"ip"`
	WSPort          int      `yaml:"wsport"`
	DatabaseType    string   `yaml:"database_type"`
	DatabaseURL     string   `yaml:"database_url"`
	DatabasePath    string   `yaml:"database_path"`
	JWTKey          string   `yaml:"jwt_key"`
	JWTAlgorithm    string   `yaml:"jwt_algorithm"`
	JWTExpiry       string   `yaml:"jwt_expiry"`
	AdminPassword   string   `yaml:"admin_password"`
	GameVersions    []string `yaml:"game_versions"`
	LibraryJSONPath string   `yaml:"library_json_path"`
	ItemsJSONPath   string   `yaml:"items_json_path"`
}

var cfg *Config
var FirstRun bool
var configMu sync.Mutex

var (
	Host            string
	Port            int
	WSPort          int
	DatabaseType    string
	DatabaseURL     string
	DatabasePath    string
	JWTKey          []byte
	JWTAlgorithm    string
	JWTExpiry       time.Duration
	AdminPassword   string
	GameVersions    []string
	LibraryJSONPath string
	ItemsJSONPath   string
)

const (
	defaultPort          = 5231
	defaultIp            = "127.0.0.1"
	defaultWSPort        = 5232
	defaultDatabaseType  = "mysql"
	defaultDatabaseURL   = "root:1234567890@tcp(127.0.0.1:3306)/users?charset=utf8mb4&parseTime=True&loc=Local"
	defaultDatabasePath  = "data/kards.db"
	defaultJWTKey        = "CometKards-is-a-help-much-kards-players-that-can't-find-gameuser-or-baned"
	defaultJWTAlgorithm  = "HS256"
	defaultJWTExpiry     = "24h"
	defaultAdminPassword = "change-this-password"
)

var defaultGameVersions = []string{
	"Kards 1.47",
	"Kards 1.49",
	"Kards 1.50",
	"Kards 1.51",
	"Kards 1.52",
	"Kards 1.53",
	"Kards 1.54",
}

func init() {
	var err error
	cfg, err = LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	Host = cfg.Ip
	Port = cfg.Port
	WSPort = cfg.WSPort
	DatabaseType = cfg.DatabaseType
	DatabaseURL = cfg.DatabaseURL
	DatabasePath = cfg.DatabasePath
	JWTKey = []byte(cfg.JWTKey)
	JWTAlgorithm = cfg.JWTAlgorithm
	AdminPassword = cfg.AdminPassword
	GameVersions = append([]string(nil), cfg.GameVersions...)
	LibraryJSONPath = cfg.LibraryJSONPath
	ItemsJSONPath = cfg.ItemsJSONPath
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
	existingKeys := loadConfigKeys()
	_, configExists := existingKeys[""]

	applyEnvOverrides(cfgFromFile)

	applyDefaults(cfgFromFile)

	// 濡傛灉閰嶇疆鏂囦欢涓嶅瓨鍦紝鍒欏垱寤哄畠
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		FirstRun = true
		if err := saveYAMLConfig(cfgFromFile); err != nil {
			return nil, fmt.Errorf("failed to create config.yaml: %w", err)
		}
	} else if configExists && shouldPersistDefaultedConfig(existingKeys) {
		if err := saveYAMLConfig(cfgFromFile); err != nil {
			return nil, fmt.Errorf("failed to update config.yaml defaults: %w", err)
		}
	}

	return cfgFromFile, nil
}

func loadYAMLConfig(cfg *Config) error {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}
	return nil
}

func saveYAMLConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0644)
}

func getConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}
	return "config.yaml"
}

func loadConfigKeys() map[string]struct{} {
	keys := make(map[string]struct{})
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return keys
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return keys
	}

	keys[""] = struct{}{}
	for key := range raw {
		keys[key] = struct{}{}
	}
	return keys
}

func shouldPersistDefaultedConfig(keys map[string]struct{}) bool {
	if _, ok := keys["admin_password"]; !ok && os.Getenv("ADMIN_PASSWORD") == "" {
		return true
	}
	if _, ok := keys["game_versions"]; !ok && os.Getenv("GAME_VERSIONS") == "" {
		return true
	}
	if _, ok := keys["library_json_path"]; !ok && os.Getenv("LIBRARY_JSON_PATH") == "" {
		return true
	}
	if _, ok := keys["items_json_path"]; !ok && os.Getenv("ITEMS_JSON_PATH") == "" {
		return true
	}
	return false
}

func NormalizeDatabaseType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "mysql":
		return "mysql"
	case "sqlite", "gorml":
		return "sqlite"
	default:
		return ""
	}
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
	if v := getEnv("DB_TYPE", ""); v != "" {
		cfg.DatabaseType = v
	}
	if v := getEnv("DB_URL", ""); v != "" {
		cfg.DatabaseURL = v
	}
	if v := getEnv("DB_PATH", ""); v != "" {
		cfg.DatabasePath = v
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
	if v := getEnv("GAME_VERSIONS", ""); v != "" {
		cfg.GameVersions = splitCSV(v)
	}
	if v := getEnv("LIBRARY_JSON_PATH", ""); v != "" {
		cfg.LibraryJSONPath = v
	}
	if v := getEnv("ITEMS_JSON_PATH", ""); v != "" {
		cfg.ItemsJSONPath = v
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
	cfg.DatabaseType = NormalizeDatabaseType(cfg.DatabaseType)
	if cfg.DatabaseType == "" {
		cfg.DatabaseType = defaultDatabaseType
	}
	if cfg.DatabaseType == "mysql" && cfg.DatabaseURL == "" {
		cfg.DatabaseURL = defaultDatabaseURL
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = defaultDatabasePath
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
	if len(cfg.GameVersions) == 0 {
		cfg.GameVersions = append([]string(nil), defaultGameVersions...)
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func PromptInitialSetup() error {
	if !FirstRun && cfg.DatabaseType != "" && cfg.Ip != "" && cfg.Port != 0 {
		return nil
	}

	fmt.Println("First-time setup detected. Please configure server parameters.")
	fmt.Print("Database type [mysql/sqlite] (default mysql): ")
	var dbType string
	fmt.Scanln(&dbType)
	dbType = NormalizeDatabaseType(dbType)
	if dbType == "" {
		dbType = defaultDatabaseType
	}

	dbURL := cfg.DatabaseURL
	dbPath := cfg.DatabasePath
	if dbType == "sqlite" {
		dbURL = ""
		fmt.Print("SQLite path (default data/kards.db): ")
		fmt.Scanln(&dbPath)
		if dbPath == "" {
			dbPath = defaultDatabasePath
		}
	} else {
		fmt.Print("MySQL URL (user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local): ")
		fmt.Scanln(&dbURL)
		if dbURL == "" {
			dbURL = defaultDatabaseURL
		}
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

	fmt.Print("Game versions, comma-separated (default built-in): ")
	var versionsInput string
	fmt.Scanln(&versionsInput)
	gameVersions := cfg.GameVersions
	if versionsInput != "" {
		gameVersions = splitCSV(versionsInput)
	}
	if len(gameVersions) == 0 {
		gameVersions = append([]string(nil), defaultGameVersions...)
	}

	fmt.Print("Library JSON path (empty uses embedded library.json): ")
	var libraryJSONPath string
	fmt.Scanln(&libraryJSONPath)

	fmt.Print("Items JSON path (empty uses embedded items_library.json): ")
	var itemsJSONPath string
	fmt.Scanln(&itemsJSONPath)

	cfg.DatabaseType = dbType
	cfg.DatabaseURL = dbURL
	cfg.DatabasePath = dbPath
	cfg.Ip = ip
	cfg.Port = port
	cfg.WSPort = wsPort
	cfg.GameVersions = gameVersions
	cfg.LibraryJSONPath = strings.TrimSpace(libraryJSONPath)
	cfg.ItemsJSONPath = strings.TrimSpace(itemsJSONPath)
	if cfg.AdminPassword == "" {
		cfg.AdminPassword = defaultAdminPassword
	}

	if err := saveYAMLConfig(cfg); err != nil {
		return err
	}

	Host = cfg.Ip
	Port = cfg.Port
	WSPort = cfg.WSPort
	DatabaseType = cfg.DatabaseType
	DatabaseURL = cfg.DatabaseURL
	DatabasePath = cfg.DatabasePath
	AdminPassword = cfg.AdminPassword
	GameVersions = append([]string(nil), cfg.GameVersions...)
	LibraryJSONPath = cfg.LibraryJSONPath
	ItemsJSONPath = cfg.ItemsJSONPath

	fmt.Println("Configuration saved. Please restart the server.")
	return nil
}

func GetConfigSnapshot() Config {
	configMu.Lock()
	defer configMu.Unlock()
	return *cfg
}

func UpdateDatabaseSettings(dbType, dbURL, dbPath string) (*Config, error) {
	configMu.Lock()
	defer configMu.Unlock()

	normalized := NormalizeDatabaseType(dbType)
	if normalized == "" {
		return nil, fmt.Errorf("invalid database_type: %s", dbType)
	}
	if normalized == "mysql" {
		dbURL = strings.TrimSpace(dbURL)
		if dbURL == "" {
			return nil, fmt.Errorf("database_url is required for mysql")
		}
		if dbPath == "" {
			dbPath = cfg.DatabasePath
		}
	} else {
		dbURL = ""
		dbPath = strings.TrimSpace(dbPath)
		if dbPath == "" {
			return nil, fmt.Errorf("database_path is required for sqlite")
		}
	}

	cfg.DatabaseType = normalized
	cfg.DatabaseURL = dbURL
	cfg.DatabasePath = dbPath

	if err := saveYAMLConfig(cfg); err != nil {
		return nil, err
	}

	DatabaseType = cfg.DatabaseType
	DatabaseURL = cfg.DatabaseURL
	DatabasePath = cfg.DatabasePath

	snapshot := *cfg
	return &snapshot, nil
}

func GetKardsTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}
