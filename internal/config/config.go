package config

import (
	"flag"
	"os"
)

type Config struct {
	AppAddr       string
	ResultBaseURL string
	FilePath      string
	DBConnStr     string
}

func ParseFlags() Config {
	var appAddr string
	var resultBaseURL string
	var fileStoragePath string
	var dbConnectionString string
	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&resultBaseURL, "b", "http://localhost:8080", "base url for short url")
	flag.StringVar(&fileStoragePath, "f", "filestorage.txt", "file storage path")
	flag.StringVar(&dbConnectionString, "d", "host=localhost user=myadmin password=123456 dbname=urlshortener sslmode=disable", "")
	flag.Parse()

	if envAppAddr := os.Getenv("SERVER_ADDRESS"); envAppAddr != "" {
		appAddr = envAppAddr
	}
	if envResultBaseURL := os.Getenv("BASE_URL"); envResultBaseURL != "" {
		resultBaseURL = envResultBaseURL
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	}
	if envDBConnectionStr := os.Getenv("DATABASE_DSN"); envDBConnectionStr != "" {
		dbConnectionString = envDBConnectionStr
	}

	return Config{
		AppAddr:       appAddr,
		ResultBaseURL: resultBaseURL,
		FilePath:      fileStoragePath,
		DBConnStr:     dbConnectionString,
	}
}
