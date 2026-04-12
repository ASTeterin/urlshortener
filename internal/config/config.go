package config

import (
	"flag"
	"os"
)

type Config struct {
	AppAddr       string
	ResultBaseUrl string
	FilePath      string
	DbConnStr     string
}

func ParseFlags() Config {
	var appAddr string
	var resultBaseUrl string
	var fileStoragePath string
	var dbConnectionString string
	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&resultBaseUrl, "b", "http://localhost:8080", "base url for short url")
	flag.StringVar(&fileStoragePath, "f", "filestorage.txt", "file storage path")
	flag.StringVar(&dbConnectionString, "d", "", "")
	flag.Parse()

	if envAppAddr := os.Getenv("SERVER_ADDRESS"); envAppAddr != "" {
		appAddr = envAppAddr
	}
	if envResultBaseUrl := os.Getenv("BASE_URL"); envResultBaseUrl != "" {
		resultBaseUrl = envResultBaseUrl
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	}
	if envDbConnectionStr := os.Getenv("DATABASE_DSN"); envDbConnectionStr != "" {
		dbConnectionString = envDbConnectionStr
	}

	return Config{
		AppAddr:       appAddr,
		ResultBaseUrl: resultBaseUrl,
		FilePath:      fileStoragePath,
		DbConnStr:     dbConnectionString,
	}
}
