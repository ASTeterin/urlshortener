package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ASTeterin/urlshortener/internal/service"
)

type Handler interface {
	GetURL(res http.ResponseWriter, req *http.Request)
	GetShortURL(res http.ResponseWriter, req *http.Request)
}

type handler struct {
	service service.ShortenerService
}

func NewHandler(service service.ShortenerService) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) GetURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	urlParts := strings.Split(req.URL.Path, "/")
	shortUrl := urlParts[1]
	originalUrl, err := h.service.GetOriginalUrl(shortUrl)
	if err != nil || originalUrl == nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Location", *originalUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *handler) GetShortURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
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

	originalUrl := string(bodyBytes)
	if len(originalUrl) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = url.ParseRequestURI(originalUrl)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	short := h.service.GetShortUrl(originalUrl)
	response := []byte("http://localhost:8080/" + short)
	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Content-Length", strconv.Itoa(len(response)))
	res.WriteHeader(http.StatusCreated)
	if _, err = res.Write(response); err != nil {
		log.Printf("Ошибка отправки ответа: %v", err)
	}
}
