package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Handler interface {
	GetURL(c *gin.Context)
	GetShortURL(c *gin.Context, baseUrl string)
	CheckDbConnection(c *gin.Context)
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
	service   service.ShortenerService
	dbConnStr string
}

type restApiHandler struct {
	service service.ShortenerService
}

func NewHandler(service service.ShortenerService, dbConnStr string) Handler {
	return &handler{
		service:   service,
		dbConnStr: dbConnStr,
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

func (h *handler) CheckDbConnection(c *gin.Context) {
	db, err := sql.Open("pgx", h.dbConnStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusOK)
}
