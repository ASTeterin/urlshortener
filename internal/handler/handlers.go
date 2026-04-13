package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"
	"net/url"
)

type Handler interface {
	GetURL(ctx context.Context, c *gin.Context)
	GetShortURL(ctx context.Context, c *gin.Context, baseURL string)
	CheckDbConnection(ctx context.Context, c *gin.Context)
}

type RestAPIHandler interface {
	GetShortURL(ctx context.Context, c *gin.Context, baseURL string)
	ListShortURLs(ctx context.Context, c *gin.Context, baseURL string)
}

type URLData struct {
	URL string `json:"url"`
}

type ShortURLData struct {
	ShortURL string `json:"result"`
}

type ListURLItem struct {
	URL           string `json:"original_url"`
	CorrelationId string `json:"correlation_id"`
}

type ListShortURLItem struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type handler struct {
	service service.ShortenerService
	dbConn  *sql.DB
}

type restAPIHandler struct {
	service service.ShortenerService
}

func NewHandler(service service.ShortenerService, dbConn *sql.DB) Handler {
	return &handler{
		service: service,
		dbConn:  dbConn,
	}
}

func NewRestAPIHandler(service service.ShortenerService) RestAPIHandler {
	return &restAPIHandler{
		service: service,
	}
}

func (h *handler) GetURL(ctx context.Context, c *gin.Context) {
	shortURL := c.Param("id")
	originalURL, err := h.service.GetOriginalURL(ctx, shortURL)
	if err != nil || originalURL == nil {
		c.AbortWithStatus(400)
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Location", *originalURL)
	c.Redirect(http.StatusTemporaryRedirect, *originalURL)
}

func (h *restAPIHandler) GetShortURL(ctx context.Context, c *gin.Context, baseURL string) {
	var urlData URLData
	err := c.BindJSON(&urlData)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	originalURL := urlData.URL
	if originalURL == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	_, err = url.ParseRequestURI(originalURL)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	shortURL, err := h.service.GetShortURL(ctx, originalURL)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	short := (fmt.Sprintf("%s/%s", baseURL, *shortURL))
	var responseData ShortURLData
	responseData.ShortURL = short
	response, err := json.Marshal(responseData)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusCreated, "application/json", response)
}

func (h *restAPIHandler) ListShortURLs(ctx context.Context, c *gin.Context, baseURL string) {
	var urls []ListURLItem
	err := c.BindJSON(&urls)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	urlsMap := make(map[string]string)
	for _, u := range urls {
		if u.URL == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		_, err = url.ParseRequestURI(u.URL)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		urlsMap[u.CorrelationId] = u.URL
	}

	shortURLsMap, err := h.service.ListShortURL(ctx, urlsMap)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseData := make([]ListShortURLItem, 0, len(shortURLsMap))
	for correlationId, v := range shortURLsMap {
		short := (fmt.Sprintf("%s/%s", baseURL, v))
		responseData = append(responseData, ListShortURLItem{
			CorrelationId: correlationId,
			ShortURL:      short,
		})
	}
	response, err := json.Marshal(responseData)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusCreated, "application/json", response)
}

func (h *handler) GetShortURL(ctx context.Context, c *gin.Context, baseURL string) {
	var originalURL string
	err := c.BindPlain(&originalURL)
	if err != nil || originalURL == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	_, err = url.ParseRequestURI(originalURL)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	short, err := h.service.GetShortURL(ctx, originalURL)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	response := []byte(fmt.Sprintf("%s/%s", baseURL, *short))

	c.Data(http.StatusCreated, "text/plain", response)
}

func (h *handler) CheckDbConnection(ctx context.Context, c *gin.Context) {
	if err := h.dbConn.PingContext(ctx); err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusOK)
}
