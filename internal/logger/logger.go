package logger

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		log.Info().Str("URI", c.Request.RequestURI).
			Str("Method", c.Request.Method).
			Str("Duration", duration.String()).
			Int("Status", status).
			Int("Size", size).
			Msg("Request complete")
	}
}

func LogErrorWithStack(err error, message string) {
	if err == nil {
		return
	}

	l := log.Error().Err(err).Str("message", message)
	stackStr := fmt.Sprintf("%+v", err)
	l = l.Str("stack", stackStr)

	l.Send()
}
