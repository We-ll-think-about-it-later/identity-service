package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

// LoggingMiddleware logs the incoming HTTP request & its duration using logrus.
func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)

				// Type assertion to check if the recovered value is an error
				if recErr, ok := err.(error); ok {
					logger.WithFields(logrus.Fields{
						"error": recErr,
						"stack": string(debug.Stack()),
					}).Error("panic recovered")
				} else {
					logger.WithFields(logrus.Fields{
						"recovered_value": err,
						"stack":           string(debug.Stack()),
					}).Error("panic recovered (non-error value)")
				}
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(c.Writer)
		c.Writer = wrapped
		c.Next()

		entry := logger.WithFields(logrus.Fields{
			"status":   wrapped.status,
			"method":   c.Request.Method,
			"path":     c.Request.URL.EscapedPath(),
			"duration": time.Since(start),
		})

		if wrapped.status >= 500 {
			entry.Error("")
		} else {
			entry.Info("")
		}
	}
}
