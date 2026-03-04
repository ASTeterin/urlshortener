package handler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ASTeterin/urlshortener/internal/repository"
	"github.com/ASTeterin/urlshortener/internal/service"
)

func Test_handler_GetURL(t *testing.T) {
	repo := repository.NewUrlRepository()
	shortenerService := service.NewShortenerService(repo)
	h := NewHandler(shortenerService)

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
			name:   "test invalid http-method",
			method: "GET",
			body:   "http://yandex.ru",
			want: want{
				code:        400,
				contentType: "",
				bodyLen:     0,
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
			// создаём новый Recorder
			w := httptest.NewRecorder()
			h.GetShortURL(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Len(t, string(resBody), test.want.bodyLen)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
