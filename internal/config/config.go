package config

import (
	"flag"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
)

const (
	defaultSigningKey string = "default_signing_key"
	maxWorkers               = 8
)

type Config struct {
	AppAddr       string
	ResultBaseURL string
	FilePath      string
	DBConnStr     string
	SigningKey    string
	MaxWorkers    int
}

func ParseFlags() Config {
	var appAddr string
	var resultBaseURL string
	var fileStoragePath string
	var dbConnectionString string
	var signingKey = defaultSigningKey
	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&resultBaseURL, "b", "http://localhost:8080", "base url for short url")
	flag.StringVar(&fileStoragePath, "f", "filestorage.txt", "file storage path")
	flag.StringVar(&dbConnectionString, "d", "", "database DSN")
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
	countWorkers := getEnvInt("MAX_WORKERS", maxWorkers)

	return Config{
		AppAddr:       appAddr,
		ResultBaseURL: resultBaseURL,
		FilePath:      fileStoragePath,
		DBConnStr:     dbConnectionString,
		SigningKey:    signingKey,
		MaxWorkers:    countWorkers,
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
