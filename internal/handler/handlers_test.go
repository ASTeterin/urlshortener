package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ASTeterin/urlshortener/internal/cookie"

	"github.com/gin-gonic/gin"

	"github.com/ASTeterin/urlshortener/internal/audit"
	"github.com/ASTeterin/urlshortener/internal/repository/file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ASTeterin/urlshortener/internal/service"
)

const (
	baseURL = "http://localhost:8000"
)

func Test_handler_GetShortURL(t *testing.T) {
	router := setupRouter("test_get_short_url")

	type want struct {
		code        int
		response    string
		contentType string
		bodyLen     int
	}

	tests := []struct {
		name   string
		method string
		body   string
		want   want
	}{
		{
			name:   "positive test",
			method: "POST",
			body:   "http://yandex.ru",
			want: want{
				code:        201,
				contentType: "text/plain",
				bodyLen:     30,
			},
		},
		{
			name:   "positive test duplicate url",
			method: "POST",
			body:   "http://yandex.ru",
			want: want{
				code:        409,
				contentType: "text/plain",
				bodyLen:     30,
			},
		},
		{
			name:   "test empty request",
			method: "POST",
			body:   "",
			want: want{
				code:        400,
				contentType: "",
				bodyLen:     0,
			},
		},
		{
			name:   "test invalid URL in request",
			method: "POST",
			body:   "123",
			want: want{
				code:        400,
				contentType: "",
				bodyLen:     0,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.body)
			request := httptest.NewRequest(test.method, "/", body)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Len(t, string(resBody), test.want.bodyLen)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func Test_restAPIHandler_GetShortURL(t *testing.T) {
	router := setupRouter("test_rest_api_get_short_url")
	type want struct {
		code        int
		response    ShortURLData
		contentType string
		hasError    bool
	}

	tests := []struct {
		name   string
		method string
		body   URLData
		want   want
	}{
		{
			name:   "positive test",
			method: "POST",
			body: URLData{
				URL: "http://ya.ru",
			},
			want: want{
				code:        201,
				contentType: "application/json",
				hasError:    false,
			},
		},
		{
			name:   "positive test duplicate url",
			method: "POST",
			body: URLData{
				URL: "http://ya.ru",
			},
			want: want{
				code:        409,
				contentType: "application/json",
				hasError:    false,
			},
		},
		{
			name:   "test empty request",
			method: "POST",
			body: URLData{
				URL: "",
			},
			want: want{
				code:        400,
				hasError:    true,
				contentType: "",
			},
		},
		{
			name:   "test invalid URL in request",
			method: "POST",
			body: URLData{
				URL: "123",
			},
			want: want{
				code:        400,
				hasError:    true,
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonData, err := json.Marshal(test.body)
			body := bytes.NewReader(jsonData)
			request := httptest.NewRequest(test.method, "/api/shorten", body)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			require.NoError(t, err)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			if !test.want.hasError {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				var shortURLData ShortURLData
				err = json.Unmarshal(resBody, &shortURLData)
				require.NoError(t, err)
			}
		})
	}
}

func Test_handler_GetURL(t *testing.T) {
	const originalURL = "http://test.ru"
	router := setupRouter("test_get_url")
	body := strings.NewReader(originalURL)
	request := httptest.NewRequest(http.MethodPost, "/", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	res := w.Result()
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	defer res.Body.Close()
	response, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	urlParts := strings.Split(string(response), "/")
	shortURL := urlParts[len(urlParts)-1]

	type want struct {
		code        int
		contentType string
		location    string
	}

	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "positive test",
			method: "GET",
			url:    shortURL,
			want: want{
				code:        307,
				contentType: "text/plain",
				location:    originalURL,
			},
		},
		{
			name:   "original url not found",
			method: "GET",
			url:    "123",
			want: want{
				code:        400,
				contentType: "",
				location:    "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/"+test.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			require.NoError(t, err)
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func Example_handler_GetURL() {
	router := setupRouter("test_example_get_url")

	body := []byte("https://example.com")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fullURL := strings.TrimSpace(w.Body.String())
	parts := strings.Split(fullURL, "/")
	shortCode := parts[len(parts)-1]

	req, _ = http.NewRequest("GET", "/"+shortCode, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 307
}

func Example_handler_GetShortURL() {
	router := setupRouter("test_example_get_short_url")

	body := []byte("https://example.com/very/long/url")
	// Путь "/" соответствует регистрации в setupRouter
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 201
}

func Example_restAPIHandler_GetShortURL() {
	router := setupRouter("rest_api_example_get_short_url")

	body := []byte(`{"url": "https://example.com/very/long/url"}`)
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 201
}

func Example_restAPIHandlerImpl_ListShortURLs() {
	router := setupRouter("rest_api_example_list_short_urls")

	body := []byte(`[
		{"original_url": "https://example.com/1", "correlation_id": "req-1"},
		{"original_url": "https://example.com/2", "correlation_id": "req-2"}
	]`)
	req, _ := http.NewRequest("POST", "/api/list", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 201
}

func Example_restAPIHandlerImpl_ListUserURLs() {
	router := setupRouter("rest_api_example_get_short_url")

	createBody := []byte(`{"url": "https://example.com/user-link"}`)
	createReq, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	req, _ := http.NewRequest("GET", "/api/user/urls", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

func Example_restAPIHandlerImpl_BatchRemove() {
	router := setupRouter("test_rest_api_batch_remove")

	createBody := []byte(`{"url": "https://example.com/to-delete"}`)
	createReq, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var response ShortURLData
	if err := json.Unmarshal(createW.Body.Bytes(), &response); err == nil && response.ShortURL != "" {
		// Извлекаем только код из полного URL
		parts := strings.Split(response.ShortURL, "/")
		shortCode := parts[len(parts)-1]

		body := []byte(fmt.Sprintf(`["%s"]`, shortCode))
		req, _ := http.NewRequest("POST", "/api/batch/remove", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		fmt.Println(w.Code)
	} else {
		panic(err)
	}
	// Output: 202
}

func setupRouter(storageFile string) *gin.Engine {
	repo, err := file.NewURLRepository(storageFile)
	if err != nil {
		panic(err)
	}
	shortenerService := service.NewShortenerService(repo, 8)

	mngr := audit.NewAuditManager()
	h := NewHandler(shortenerService, nil, mngr)
	restAPIHandler := NewRestAPIHandler(shortenerService)

	r := gin.Default()

	testAuthMiddleware := func(c *gin.Context) {
		if c.GetString(cookie.GetUserKey()) == "" {
			c.Set(cookie.GetUserKey(), "test-user-123")
		}
		c.Next()
	}

	r.GET("/:id", h.GetURL)
	r.POST("/", func(c *gin.Context) {
		h.GetShortURL(c, baseURL)
	})
	r.POST("/api/shorten", testAuthMiddleware, func(c *gin.Context) {
		restAPIHandler.GetShortURL(c, baseURL)
	})
	r.POST("/api/list", testAuthMiddleware, func(c *gin.Context) {
		restAPIHandler.ListShortURLs(c, baseURL)
	})
	r.GET("/api/user/urls", testAuthMiddleware, func(c *gin.Context) {
		restAPIHandler.ListUserURLs(c, baseURL)
	})
	r.POST("/api/batch/remove", testAuthMiddleware, func(c *gin.Context) {
		restAPIHandler.BatchRemove(c)
	})

	return r
}
