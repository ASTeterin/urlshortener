package config

import "flag"

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
	return Config{
		AppAddr:       appAddr,
		ResultBaseUrl: resultBaseUrl,
	}
}
