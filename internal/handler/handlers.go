package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

var URL string

func Handle(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		handlePostRequest(res, req)
	} else if req.Method == "GET" {
		handleGetRequest(res, req)
	}
}

func handleGetRequest(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	res.Header().Add("Location", URL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func handlePostRequest(res http.ResponseWriter, req *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(req.Body)

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	URL = string(bodyBytes)
	if len(URL) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = url.ParseRequestURI(URL)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	response := []byte("http://localhost:8080/EwHXdJf")
	res.Header().Set("Content-Type", "text/plain")
	res.Header().Add("Content-Length", strconv.Itoa(len(response)))
	res.WriteHeader(http.StatusCreated)
	if _, err = res.Write([]byte(response)); err != nil {
		log.Printf("Ошибка отправки ответа: %v", err)
	}
}
