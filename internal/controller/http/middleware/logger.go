package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/We-ll-think-about-it-later/identity-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	gin.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w gin.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// LoggingMiddleware logs the incoming HTTP request & its duration using zerolog.
func LoggingMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)

				// Type assertion to check if the recovered value is an error
				if recErr, ok := err.(error); ok {
					logger.Zerolog.Error().
						Err(recErr). // Pass the error if it is one
						Bytes("stack", debug.Stack()).
						Msg("panic recovered")
				} else {
					logger.Zerolog.Error().
						Interface("recovered_value", err). // Log the raw value otherwise
						Bytes("stack", debug.Stack()).
						Msg("panic recovered (non-error value)")
				}
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(c.Writer)
		c.Writer = wrapped
		c.Next()

		var logEvent *zerolog.Event
		if wrapped.status >= 500 {
			logEvent = logger.Zerolog.Error()
		} else {
			logEvent = logger.Zerolog.Info()
		}

		logEvent.
			Int("status", wrapped.status).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.EscapedPath()).
			Dur("duration", time.Since(start)).
			Msg("")
	}
}
