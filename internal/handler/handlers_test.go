package handler

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ASTeterin/urlshortener/internal/repository"
	"github.com/ASTeterin/urlshortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseUrl = "localhost:8000"

func Test_handler_GetShortURL(t *testing.T) {
	router := setupRouter()

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

func Test_handler_GetURL(t *testing.T) {
	const originalUrl = "http://yandex.ru"
	router := setupRouter()
	body := strings.NewReader(originalUrl)
	request := httptest.NewRequest(http.MethodPost, "/", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	res := w.Result()
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	defer res.Body.Close()
	response, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	urlParts := strings.Split(string(response), "/")
	shortUrl := urlParts[len(urlParts)-1]

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
			url:    shortUrl,
			want: want{
				code:        307,
				contentType: "text/plain",
				location:    originalUrl,
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

			require.NoError(t, err)
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func setupRouter() *gin.Engine {
	repo := repository.NewUrlRepository()
	shortenerService := service.NewShortenerService(repo)
	h := NewHandler(shortenerService)

	r := gin.Default()
	r.POST("/", func(c *gin.Context) {
		h.GetShortURL(c, baseUrl)
	})
	r.GET("/:id", func(c *gin.Context) {
		h.GetURL(c)
	})

	return r
}
