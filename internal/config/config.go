package config

import (
	"flag"
	"os"
)

type Config struct {
	AppAddr       string
	ResultBaseUrl string
}

func ParseFlags() Config {
	var appAddr string
	var resultBaseUrl string
	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&resultBaseUrl, "b", "http://localhost:8080", "base url for short url")
	flag.Parse()

	if envAppAddr := os.Getenv("SERVER_ADDRESS"); envAppAddr != "" {
		appAddr = envAppAddr
	}
	if envResultBaseUrl := os.Getenv("BASE_URL"); envResultBaseUrl != "" {
		resultBaseUrl = envResultBaseUrl
	}

	return Config{
		AppAddr:       appAddr,
		ResultBaseUrl: resultBaseUrl,
	}
}
