//go:generate go run ../../cmd/reset/main.go
package config

import (
	"errors"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultPort            string = ":8080"
	grpcPort               string = ":8081"
	defaultBaseUrl         string = "http://localhost:8080"
	defaultFileStoragePath string = "filestorage.txt"
	maxWorkers             int    = 8
	certFile               string = "cert.pem"
	keyFile                string = "key.pem"
)

var (
	errSigningKeyNotSet = errors.New("signing key not set")
)

// generate:reset
type Config struct {
	ServerAddr    string `json:"server_addr" mapstructure:"server_addr"`
	GRPCAddr      string `json:"grpc_addr" mapstructure:"grpc_addr"`
	ResultBaseURL string `json:"result_base_url" mapstructure:"result_base_url"`
	FilePath      string `json:"file_path" mapstructure:"file_path"`
	DatabaseURL   string `json:"database_url" mapstructure:"database_url"`
	SigningKey    string `json:"signing_key" mapstructure:"signing_key"`
	MaxWorkers    int    `json:"max_workers" mapstructure:"max_workers"`
	AuditFilePath string `json:"audit_file_path" mapstructure:"audit_file_path"`
	AuditURL      string `json:"audit_url" mapstructure:"audit_url"`
	EnableHTTPS   bool   `json:"enable_https" mapstructure:"enable_https"`
	CertFile      string `json:"cert_file" mapstructure:"cert_file"`
	KeyFile       string `json:"key_file" mapstructure:"key_file"`
	TrustedSubnet string `json:"trusted_subnet" mapstructure:"trusted_subnet"`
}

func (c *Config) Validate() error {
	if c.SigningKey == "" {
		return errSigningKeyNotSet
	}
	return nil
}

func ParseFlags() Config {
	var configPath string

	pflag.StringVarP(&configPath, "config", "c", "", "path to config file")
	pflag.StringVarP(nil, "a", "a", "", "port to run server")
	pflag.StringVarP(nil, "b", "b", "", "base url for short url")
	pflag.StringVarP(nil, "f", "f", "", "file storage path")
	pflag.StringVarP(nil, "d", "d", "", "database DSN")
	pflag.StringVarP(nil, "audit-file", "", "", "audit file path")
	pflag.StringVarP(nil, "audit-url", "", "", "audit service url")
	pflag.BoolVarP(nil, "s", "s", false, "enable HTTPS")
	pflag.StringVar(nil, "t", "t", "CIDR subnet for /api/internal/stats")
	pflag.Parse()

	viper.SetConfigType("json")
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		viper.SetConfigFile(envConfig)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.SetDefault("server_addr", defaultPort)
	viper.SetDefault("grpc_addr", grpcPort)
	viper.SetDefault("result_base_url", defaultBaseUrl)
	viper.SetDefault("file_path", defaultFileStoragePath)
	viper.SetDefault("database_url", "")
	viper.SetDefault("signing_key", "")
	viper.SetDefault("max_workers", maxWorkers)
	viper.SetDefault("audit_file_path", "")
	viper.SetDefault("audit_url", "")
	viper.SetDefault("enable_https", false)
	viper.SetDefault("cert_file", certFile)
	viper.SetDefault("key_file", keyFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Warn().Err(err).Msg("failed to read config file")
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindEnv("server_addr", "SERVER_ADDRESS")
	viper.BindEnv("grpc_addr", "GRPC_ADDRESS")
	viper.BindEnv("result_base_url", "BASE_URL")
	viper.BindEnv("file_path", "FILE_STORAGE_PATH")
	viper.BindEnv("database_url", "DATABASE_DSN")
	viper.BindEnv("signing_key", "SIGNING_KEY")
	viper.BindEnv("audit_file_path", "AUDIT_FILE")
	viper.BindEnv("audit_url", "AUDIT_URL")
	viper.BindEnv("enable_https", "ENABLE_HTTPS")
	viper.BindEnv("max_workers", "MAX_WORKERS")
	viper.BindEnv("trusted_subnet", "TRUSTED_SUBNET")

	viper.BindPFlag("server_addr", pflag.Lookup("a"))
	viper.BindPFlag("result_base_url", pflag.Lookup("b"))
	viper.BindPFlag("file_path", pflag.Lookup("f"))
	viper.BindPFlag("database_url", pflag.Lookup("d"))
	viper.BindPFlag("audit_file_path", pflag.Lookup("audit-file"))
	viper.BindPFlag("audit_url", pflag.Lookup("audit-url"))
	viper.BindPFlag("enable_https", pflag.Lookup("s"))
	viper.BindPFlag("trusted_subnet", pflag.Lookup("t"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal config")
	}

	return cfg
}
