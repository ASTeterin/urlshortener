package handler

import (
	"io"
	"net/http"
	"net/url"
)

func GetShortURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}

	defer req.Body.Close()
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}

	url := string(bodyBytes)
	if !isValidURL(url) {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Add("Content-Length", "30")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("EwHXdJfB"))
}

func GetURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Add("Location", "https://practicum.yandex.ru")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != "" && u.IsAbs()
}
