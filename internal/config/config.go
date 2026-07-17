//go:generate go run ../../cmd/reset/main.go
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	defaultPort            string = ":8080"
	defaultBaseUrl         string = "http://localhost:8080"
	defaultFileStoragePath        = "filestorage.txt"
	defaultSigningKey      string = "default_signing_key"
	maxWorkers                    = 8
	certFile                      = "cert.pem"
	keyFile                       = "key.pem"
)

var (
	fAppAddr, fResultBaseURL, fFilePath, fDBConn, fAuditFile, fAuditURL, fConfigPath string
	fEnableHTTPS                                                                     bool
)

// generate:reset
type Config struct {
	ServerAddr    string `json:"server_addr"`
	ResultBaseURL string `json:"result_base_url"`
	FilePath      string `json:"file_path"`
	DatabaseURL   string `json:"database_url"`
	SigningKey    string `json:"signing_key"`
	MaxWorkers    int    `json:"max_workers"`
	AuditFilePath string `json:"audit_file_path"`
	AuditURL      string `json:"audit_url"`
	EnableHTTPS   bool   `json:"enable_https"`
	CertFile      string `json:"cert_file"`
	KeyFile       string `json:"key_file"`
}

func ParseFlags() Config {
	// 1. Базовые значения
	cfg := Config{
		ServerAddr:    defaultPort,
		ResultBaseURL: defaultBaseUrl,
		FilePath:      defaultFileStoragePath,
		DatabaseURL:   "",
		SigningKey:    defaultSigningKey,
		MaxWorkers:    maxWorkers,
		AuditFilePath: "",
		AuditURL:      "",
		EnableHTTPS:   false,
		CertFile:      certFile,
		KeyFile:       keyFile,
	}

	// 2. Регистрируем флаги с пустыми дефолтами
	flag.StringVar(&fAppAddr, "a", "", "port to run server")
	flag.StringVar(&fResultBaseURL, "b", "", "base url for short url")
	flag.StringVar(&fFilePath, "f", "", "file storage path")
	flag.StringVar(&fDBConn, "d", "", "database DSN")
	flag.StringVar(&fAuditFile, "audit-file", "", "audit file path")
	flag.StringVar(&fAuditURL, "audit-url", "", "audit service url")
	flag.BoolVar(&fEnableHTTPS, "s", false, "enable HTTPS")
	flag.StringVar(&fConfigPath, "config", "", "path to config file")
	flag.StringVar(&fConfigPath, "c", "", "shorthand for -config")
	flag.Parse()

	// 3. Загрузка JSON-конфига (низший приоритет)
	configPath := fConfigPath
	if configPath == "" {
		if p, exists := os.LookupEnv("CONFIG"); exists {
			configPath = p
		}
	}
	if configPath != "" {
		if err := loadConfigFile(&cfg, configPath); err != nil {
			log.Warn().Err(err).Msg("failed to load config file, skipping")
		}
	}

	// 4. Переменные окружения (средний приоритет)
	applyEnv(&cfg)

	// 5. Флаги (высший приоритет)
	applyFlags(&cfg)

	return cfg
}

func loadConfigFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, cfg)
}

func applyEnv(cfg *Config) {
	if v, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		cfg.ServerAddr = v
	}
	if v, exists := os.LookupEnv("BASE_URL"); exists {
		cfg.ResultBaseURL = v
	}
	if v, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		cfg.FilePath = v
	}
	if v, exists := os.LookupEnv("DATABASE_DSN"); exists {
		cfg.DatabaseURL = v
	}
	if v, exists := os.LookupEnv("SIGNING_KEY"); exists {
		cfg.SigningKey = v
	}
	if v, exists := os.LookupEnv("AUDIT_FILE"); exists {
		cfg.AuditFilePath = v
	}
	if v, exists := os.LookupEnv("AUDIT_URL"); exists {
		cfg.AuditURL = v
	}
	if v, exists := os.LookupEnv("ENABLE_HTTPS"); exists {
		cfg.EnableHTTPS = v == "true" || v == "1"
	}
	if v, exists := os.LookupEnv("MAX_WORKERS"); exists {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxWorkers = n
		} else {
			log.Info().Str("err", err.Error()).Str("key", "MAX_WORKERS").Str("value", v).Msg("failed to parse")
		}
	}
}

func applyFlags(cfg *Config) {
	if fAppAddr != "" {
		cfg.ServerAddr = fAppAddr
	}
	if fResultBaseURL != "" {
		cfg.ResultBaseURL = fResultBaseURL
	}
	if fFilePath != "" {
		cfg.FilePath = fFilePath
	}
	if fDBConn != "" {
		cfg.DatabaseURL = fDBConn
	}
	if fAuditFile != "" {
		cfg.AuditFilePath = fAuditFile
	}
	if fAuditURL != "" {
		cfg.AuditURL = fAuditURL
	}

	if fEnableHTTPS {
		cfg.EnableHTTPS = true
	}
}
