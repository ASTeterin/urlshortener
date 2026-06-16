package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ASTeterin/urlshortener/internal/audit"
	"github.com/ASTeterin/urlshortener/internal/cookie"
	"github.com/ASTeterin/urlshortener/internal/logger"
	"github.com/ASTeterin/urlshortener/internal/model"
	"github.com/ASTeterin/urlshortener/internal/service"
)

// Handler defines the HTTP handlers for the main URL shortener API.
type Handler interface {
	// GetURL redirects to the original URL based on the short code.
	// Returns 307 Temporary Redirect on success, 410 Gone if deleted, 400 Bad Request otherwise.
	GetURL(c *gin.Context)

	// GetShortURL creates a short URL from a plain text URL in the request body.
	// Returns 201 Created with the full short URL, 409 Conflict if duplicate, 400/500 on error.
	GetShortURL(c *gin.Context, baseURL string)

	// CheckDBConnection verifies the database connectivity.
	// Returns 200 OK on success, 500 Internal Server Error on failure.
	CheckDBConnection(c *gin.Context)
}

// RestAPIHandler defines the HTTP handlers for the REST API.
type RestAPIHandler interface {
	// GetShortURL creates a short URL from a JSON payload containing the original URL.
	// Returns 201 Created or 409 Conflict with JSON response, 400 on validation error.
	GetShortURL(c *gin.Context, baseURL string)

	// ListShortURLs resolves a batch of original URLs to their short counterparts.
	// Returns 201 Created with JSON array mapping correlation IDs to short URLs, 400 on error.
	ListShortURLs(c *gin.Context, baseURL string)

	// ListUserURLs retrieves all short URLs created by the authenticated user.
	// Returns 200 OK with JSON array, 204 No Content if empty, 400/500 on error.
	ListUserURLs(c *gin.Context, baseURL string)

	// BatchRemove initiates asynchronous deletion of a list of short URLs.
	// Returns 202 Accepted immediately. Errors are logged asynchronously.
	BatchRemove(c *gin.Context)
}

// URLData represents the JSON request body for creating a short URL.
type URLData struct {
	URL string `json:"url"`
}

// ShortURLData represents the JSON response for a created short URL.
type ShortURLData struct {
	ShortURL string `json:"result"`
}

// ListURLItem represents an item in the batch request for listing short URLs.
type ListURLItem struct {
	URL           string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

// ListShortURLItem represents an item in the batch response for listing short URLs.
type ListShortURLItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ListUserURLItem represents a user's URL mapping in the response.
type ListUserURLItem struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type handler struct {
	service      service.ShortenerService
	dbConn       *sql.DB
	auditManager *audit.Manager
}

type restAPIHandler struct {
	service      service.ShortenerService
	auditManager *audit.Manager
}

// NewHandler creates a new instance of the main HTTP handler.
func NewHandler(service service.ShortenerService, dbConn *sql.DB, mngr *audit.Manager) Handler {
	return &handler{
		service:      service,
		dbConn:       dbConn,
		auditManager: mngr,
	}
}

// NewRestAPIHandler creates a new instance of the REST API HTTP handler.
func NewRestAPIHandler(service service.ShortenerService, mngr *audit.Manager) RestAPIHandler {
	return &restAPIHandler{
		service:      service,
		auditManager: mngr,
	}
}

// GetURL redirects to the original URL based on the short code.
// Returns 307 Temporary Redirect on success, 410 Gone if deleted, 400 Bad Request otherwise.
func (h *handler) GetURL(c *gin.Context) {
	shortURL := c.Param("id")
	userID := getUserID(c)
	originalURL, err := h.service.GetOriginalURL(shortURL)
	if err != nil || originalURL == nil {
		if errors.Is(err, model.ErrURLHasBeenDeleted) {
			c.AbortWithStatus(http.StatusGone)
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	notify(h.auditManager, userID, *originalURL, audit.Follow)

	c.Header("Content-Type", "text/plain")
	c.Header("Location", *originalURL)
	c.Redirect(http.StatusTemporaryRedirect, *originalURL)
}

// GetShortURL creates a short URL from a JSON payload containing the original URL.
// Returns 201 Created or 409 Conflict with JSON response, 400 on validation error.
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

	userID := getUserID(c)
	shortURL, err := h.service.GetShortURL(originalURL, userID)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateURL) {
			returnResponseWithStatus(c, http.StatusConflict, baseURL, *shortURL)
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	notify(h.auditManager, userID, originalURL, audit.Shorten)

	returnResponseWithStatus(c, http.StatusCreated, baseURL, *shortURL)
}

// ListUserURLs retrieves all short URLs created by the authenticated user.
// Returns 200 OK with JSON array, 204 No Content if empty, 400/500 on error.
func (h *restAPIHandler) ListUserURLs(c *gin.Context, baseURL string) {
	userID := getUserID(c)
	shortURLsMap, err := h.service.ListUserURLs(userID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if len(shortURLsMap) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	responseData := make([]ListUserURLItem, 0, len(shortURLsMap))
	for originalURL, shortURL := range shortURLsMap {
		short, err2 := url.JoinPath(baseURL, shortURL)
		if err2 != nil {
			logger.LogErrorWithStack(err2, "Failed to join URL path")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		responseData = append(responseData, ListUserURLItem{
			OriginalURL: originalURL,
			ShortURL:    short,
		})
	}

	response, err := json.Marshal(responseData)
	if err != nil {
		logger.LogErrorWithStack(err, "Failed to serialize JSON response")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", response)
}

// BatchRemove initiates asynchronous deletion of a list of short URLs.
// Returns 202 Accepted immediately. Errors are logged asynchronously.
func (h *restAPIHandler) BatchRemove(c *gin.Context) {
	var urls []string
	err := c.BindJSON(&urls)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	go func() {
		h.processBatchRemoveAsync(urls)
	}()

	c.Status(http.StatusAccepted)
}

// ListShortURLs resolves a batch of original URLs to their short counterparts.
// Returns 201 Created with JSON array mapping correlation IDs to short URLs, 400 on error.
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

	userID := getUserID(c)
	shortURLsMap, err := h.service.ListShortURL(urlsMap, userID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	responseData := make([]ListShortURLItem, 0, len(shortURLsMap))
	for correlationID, shortURL := range shortURLsMap {
		short, err2 := url.JoinPath(baseURL, shortURL)
		if err2 != nil {
			logger.LogErrorWithStack(err2, "Failed to join URL path")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		responseData = append(responseData, ListShortURLItem{
			CorrelationID: correlationID,
			ShortURL:      short,
		})
	}
	response, err := json.Marshal(responseData)
	if err != nil {
		logger.LogErrorWithStack(err, "Failed to serialize JSON response")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusCreated, "application/json", response)
}

// GetShortURL creates a short URL from a plain text URL in the request body.
// Returns 201 Created with the full short URL, 409 Conflict if duplicate, 400/500 on error.
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

	userID := getUserID(c)
	short, err := h.service.GetShortURL(originalURL, userID)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateURL) {
			shortURL, err2 := url.JoinPath(baseURL, *short)
			if err2 != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Data(http.StatusConflict, "text/plain", []byte(shortURL))
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	notify(h.auditManager, userID, originalURL, audit.Shorten)

	shortURL, err2 := url.JoinPath(baseURL, *short)
	if err2 != nil {
		logger.LogErrorWithStack(err2, "Failed to join URL path")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusCreated, "text/plain", []byte(shortURL))
}

// CheckDBConnection verifies the database connectivity.
// Returns 200 OK on success, 500 Internal Server Error on failure.
func (h *handler) CheckDBConnection(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	if err := h.dbConn.PingContext(ctx); err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Status(http.StatusOK)
}

func (h *restAPIHandler) processBatchRemoveAsync(urls []string) {
	result := h.service.BatchRemove(urls)

	if result != nil && len(result.Errors) > 0 {
		err := errors.Join(result.Errors...)
		logger.LogErrorWithStack(err, "processing failed")
	}
}

func returnResponseWithStatus(c *gin.Context, status int, baseURL, shortURL string) {
	short, err := url.JoinPath(baseURL, shortURL)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var responseData ShortURLData
	responseData.ShortURL = short
	response, err := json.Marshal(responseData)
	if err != nil {
		logger.LogErrorWithStack(err, "Failed to serialize JSON response")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(status, "application/json", response)
}

func getUserID(c *gin.Context) string {
	return c.GetString(cookie.GetUserKey())
}

func notify(mngr *audit.Manager, userID, originalURL string, action audit.Action) {
	event := audit.Event{
		TS:     time.Now().Unix(),
		Action: string(action),
		UserID: userID,
		URL:    originalURL,
	}
	mngr.Notify(event)
}
