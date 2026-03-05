package handler

import (
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

type handler struct {
	service service.ShortenerService
}

func NewHandler(service service.ShortenerService) Handler {
	return &handler{
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

func (h *handler) GetShortURL(c *gin.Context, baseUrl string) {
	var originalUrl string
	err := c.BindPlain(&originalUrl)
	if err != nil || originalUrl == "" {
		c.AbortWithStatus(400)
		return
	}

	_, err = url.ParseRequestURI(originalUrl)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	short := h.service.GetShortUrl(originalUrl)
	response := []byte(fmt.Sprintf("%s/%s", baseUrl, short))

	c.Data(http.StatusCreated, "text/plain", response)
}
