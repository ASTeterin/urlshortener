package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ASTeterin/urlshortener/internal/model"
	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"
	"net/url"
)

type Handler interface {
	GetURL(c *gin.Context)
	GetShortURL(c *gin.Context, baseURL string)
	CheckDBConnection(c *gin.Context)
}

type RestAPIHandler interface {
	GetShortURL(c *gin.Context, baseURL string)
	ListShortURLs(c *gin.Context, baseURL string)
}

type URLData struct {
	URL string `json:"url"`
}

type ShortURLData struct {
	ShortURL string `json:"result"`
}

type ListURLItem struct {
	URL           string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type ListShortURLItem struct {
	CorrelationID string `json:"correlation_id"`
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

func (h *handler) GetURL(c *gin.Context) {
	shortURL := c.Param("id")
	originalURL, err := h.service.GetOriginalURL(shortURL)
	if err != nil || originalURL == nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Location", *originalURL)
	c.Redirect(http.StatusTemporaryRedirect, *originalURL)
}

func (h *restAPIHandler) GetShortURL(c *gin.Context, baseURL string) {
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

	shortURL, err := h.service.GetShortURL(originalURL)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateURL) {
			returnResponseWithStatus(c, http.StatusConflict, baseURL, *shortURL)
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	returnResponseWithStatus(c, http.StatusCreated, baseURL, *shortURL)
}

func (h *restAPIHandler) ListShortURLs(c *gin.Context, baseURL string) {
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
		urlsMap[u.CorrelationID] = u.URL
	}

	shortURLsMap, err := h.service.ListShortURL(urlsMap)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseData := make([]ListShortURLItem, 0, len(shortURLsMap))
	for correlationID, shortURL := range shortURLsMap {
		short := (fmt.Sprintf("%s/%s", baseURL, shortURL))
		responseData = append(responseData, ListShortURLItem{
			CorrelationID: correlationID,
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

func (h *handler) GetShortURL(c *gin.Context, baseURL string) {
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

	short, err := h.service.GetShortURL(originalURL)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateURL) {
			response := []byte(fmt.Sprintf("%s/%s", baseURL, *short))
			c.Data(http.StatusConflict, "text/plain", response)
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	response := []byte(fmt.Sprintf("%s/%s", baseURL, *short))
	c.Data(http.StatusCreated, "text/plain", response)
}

func (h *handler) CheckDBConnection(c *gin.Context) {
	if err := h.dbConn.Ping(); err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusOK)
}

func returnResponseWithStatus(c *gin.Context, status int, baseURL, shortURL string) {
	short := (fmt.Sprintf("%s/%s", baseURL, shortURL))
	var responseData ShortURLData
	responseData.ShortURL = short
	response, err := json.Marshal(responseData)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(status, "application/json", response)
}
