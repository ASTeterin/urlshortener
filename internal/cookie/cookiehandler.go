package cookie

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	cookieName = "Authorization"
	userIDKey  = "user_id"
)

var (
	cookieExpiry = 7 * 24 * time.Hour
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func CookieHandler(signingKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID string
		authHeader := c.GetHeader(cookieName)
		if authHeader != "" {
			token, err := jwt.ParseWithClaims(authHeader, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(signingKey), nil
			})

			if err == nil && token.Valid {
				claims := token.Claims.(*Claims)
				userID = claims.UserID
				c.Set(GetUserKey(), userID)
				c.Next()
				return
			}
		}

		userID = uuid.New().String()
		claims := &Claims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   userID,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(signingKey))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
			return
		}

		c.Set(GetUserKey(), userID)
		c.SetCookie(
			cookieName,
			tokenString,
			int(cookieExpiry.Seconds()),
			"/",
			"",
			true,
			true,
		)
		c.Header(cookieName, tokenString)
		c.Next()
	}
}

func GetUserKey() string {
	return userIDKey
}
