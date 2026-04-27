package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strings"
	"time"
)

var (
	CookieName   = "Authorization"
	CookieExpiry = 7 * 24 * time.Hour // 7 дней
)

type CookieData struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CookieContextKey string

const UserIDKey CookieContextKey = "user_id"

func CookieHandler(signingKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID string
		var cookieValue string

		authHeader := c.GetHeader(CookieName)
		if authHeader != "" && verifySignedCookie(authHeader, signingKey) {
			data, err := parseCookieData(authHeader)
			if err == nil && !isCookieExpired(data) {
				userID = data.UserID
			}
		}

		if userID == "" {
			userID = uuid.New().String()
			cookieValue = generateSignedCookie(userID, signingKey)
		}

		c.Set(string(UserIDKey), userID)
		c.SetCookie(
			CookieName,
			cookieValue,
			int(CookieExpiry.Seconds()),
			"/",
			"",
			true,
			true,
		)
		c.Header(CookieName, cookieValue)
		c.Next()
	}
}

func generateSignedCookie(userID, signingKey string) string {
	data := CookieData{
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	dataBytes, _ := json.Marshal(data)
	dataStr := base64.URLEncoding.EncodeToString(dataBytes)

	h := hmac.New(sha256.New, []byte(signingKey))
	h.Write([]byte(dataStr))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return dataStr + "." + signature
}

func verifySignedCookie(cookieValue, signingKey string) bool {
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 2 {
		return false
	}

	dataStr, encodedSig := parts[0], parts[1]
	signature, err := base64.URLEncoding.DecodeString(encodedSig)
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, []byte(signingKey))
	h.Write([]byte(dataStr))
	expectedSig := h.Sum(nil)

	return hmac.Equal(signature, expectedSig)
}

func parseCookieData(cookieValue string) (*CookieData, error) {
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 2 {
		return nil, nil
	}

	dataStr := parts[0]
	dataBytes, err := base64.URLEncoding.DecodeString(dataStr)
	if err != nil {
		return nil, err
	}

	var data CookieData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func isCookieExpired(data *CookieData) bool {
	return time.Since(data.CreatedAt) > CookieExpiry
}
