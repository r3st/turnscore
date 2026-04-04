package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RequestLogger logs method, path, status and latency for every request.
// Query string, user IDs and request body are intentionally omitted (GDPR / no PII).
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		event := log.Info()
		if status >= 500 && len(c.Errors) > 0 {
			event = log.Error().Str("error", c.Errors.Last().Error())
		}
		event.
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path). // no RawQuery — may contain codes
			Int("status", status).
			Dur("latency", time.Since(start)).
			Msg("request")
	}
}
