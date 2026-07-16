//go:generate go run ../../cmd/reset/main.go
package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	defaultSigningKey string = "default_signing_key"
	maxWorkers               = 8
	certFile                 = "cert.pem"
	keyFile                  = "key.pem"
)

// generate:reset
type Config struct {
	ServerAddr    string
	ResultBaseURL string
	FilePath      string
	DatabaseURL   string
	SigningKey    string
	MaxWorkers    int
	AuditFilePath string
	AuditURL      string
	EnableHTTPS   bool
	CertFile      string
	KeyFile       string
}

func ParseFlags() Config {
	var (
		appAddr, resultBaseURL, fileStoragePath, dbConnectionString, auditFilePath, auditServiceURL string
	)
	var signingKey = defaultSigningKey
	var enableHTTPS bool

	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&resultBaseURL, "b", "http://localhost:8080", "base url for short url")
	flag.StringVar(&fileStoragePath, "f", "filestorage.txt", "file storage path")
	flag.StringVar(&dbConnectionString, "d", "", "database DSN")
	flag.StringVar(&auditFilePath, "audit-file", "", "audit file path")
	flag.StringVar(&auditServiceURL, "audit-url", "", "audit service url")
	flag.BoolVar(&enableHTTPS, "s", false, "enable HTTPS")
	flag.Parse()

	if envAppAddr, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		appAddr = envAppAddr
	}
	if envResultBaseURL, exist := os.LookupEnv("BASE_URL"); exist {
		resultBaseURL = envResultBaseURL
	}
	if envFileStoragePath, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		fileStoragePath = envFileStoragePath
	}
	if envDBConnectionStr, exist := os.LookupEnv("DATABASE_DSN"); exist {
		dbConnectionString = envDBConnectionStr
	}
	if key, exist := os.LookupEnv("SIGNING_KEY"); exist {
		signingKey = key
	}
	if envAuditFilePath, exist := os.LookupEnv("AUDIT_FILE"); exist {
		auditFilePath = envAuditFilePath
	}
	if envAuditFileURL, exist := os.LookupEnv("AUDIT_URL"); exist {
		auditServiceURL = envAuditFileURL
	}
	if envEnableHTTPS, exist := os.LookupEnv("ENABLE_HTTPS"); exist {
		enableHTTPS = envEnableHTTPS == "true" || envEnableHTTPS == "1"
	}
	countWorkers := getEnvInt("MAX_WORKERS", maxWorkers)

	return Config{
		ServerAddr:    appAddr,
		ResultBaseURL: resultBaseURL,
		FilePath:      fileStoragePath,
		DatabaseURL:   dbConnectionString,
		SigningKey:    signingKey,
		MaxWorkers:    countWorkers,
		AuditFilePath: auditFilePath,
		AuditURL:      auditServiceURL,
		EnableHTTPS:   enableHTTPS,
		CertFile:      certFile,
		KeyFile:       keyFile,
	}
}

func getEnvInt(key string, defaultValue int) int {
	valueStr, exist := os.LookupEnv(key)
	if !exist {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Info().Str("err", err.Error()).
			Str("key", key).
			Str("value", valueStr).
			Msg("failed to parse")
		return defaultValue
	}

	return value
}
