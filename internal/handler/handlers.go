package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	GetURL(c *gin.Context)
	GetShortURL(c *gin.Context, baseUrl string)
}

type RestApiHandler interface {
	GetShortURL(c *gin.Context, baseUrl string)
}

type UrlData struct {
	URL string `json:"url"`
}

type ShortUrlData struct {
	ShortUrl string `json:"result"`
}

type handler struct {
	service service.ShortenerService
}

type restApiHandler struct {
	service service.ShortenerService
}

func NewHandler(service service.ShortenerService) Handler {
	return &handler{
		service: service,
	}
}

func NewRestApiHandler(service service.ShortenerService) RestApiHandler {
	return &restApiHandler{
		service: service,
	}
}

func (h *handler) GetURL(c *gin.Context) {
	shortUrl := c.Param("id")
	originalUrl, err := h.service.GetOriginalUrl(shortUrl)
	if err != nil || originalUrl == nil {
		c.AbortWithStatus(400)
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Location", *originalUrl)
	c.Redirect(http.StatusTemporaryRedirect, *originalUrl)
}

func (h *restApiHandler) GetShortURL(c *gin.Context, baseUrl string) {
	var urlData UrlData
	err := c.BindJSON(&urlData)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	originalUrl := urlData.URL
	if originalUrl == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	_, err = url.ParseRequestURI(originalUrl)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	shortUrl, err := h.service.GetShortUrl(originalUrl)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	short := (fmt.Sprintf("%s/%s", baseUrl, *shortUrl))
	var responseData ShortUrlData
	responseData.ShortUrl = short
	response, err := json.Marshal(responseData)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusCreated, "application/json", response)

}

func (h *handler) GetShortURL(c *gin.Context, baseUrl string) {
	var originalUrl string
	err := c.BindPlain(&originalUrl)
	if err != nil || originalUrl == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	_, err = url.ParseRequestURI(originalUrl)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	short, err := h.service.GetShortUrl(originalUrl)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	response := []byte(fmt.Sprintf("%s/%s", baseUrl, *short))

	c.Data(http.StatusCreated, "text/plain", response)
}
