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
	flag.StringVar(&dbConnectionString, "d", "", "")
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

	return Config{
		AppAddr:       appAddr,
		ResultBaseURL: resultBaseURL,
		FilePath:      fileStoragePath,
		DBConnStr:     dbConnectionString,
	}
}
