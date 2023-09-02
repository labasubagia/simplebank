package api

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {

	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		duration := time.Since(startTime)

		logger := log.Info()
		if c.Writer.Status() >= 500 {
			bytes, _ := io.ReadAll(c.Request.Body)
			logger = log.Error().Bytes("body", bytes)
		}

		logger.
			Str("protocol", "http").
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status_code", c.Writer.Status()).
			Str("status_text", http.StatusText(c.Writer.Status())).
			Dur("duration", duration).
			Msg("received http request")
	}
}
