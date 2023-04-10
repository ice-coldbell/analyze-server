package receiver

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
)

func (c *core) setRequsetLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("logger", c.l.Named("REQUEST"))
	}
}

func getLogger(ctx *gin.Context) logger.Logger {
	return ctx.MustGet("logger").(logger.Logger)
}

// Support : gin recovery
// See https://github.com/gin-gonic/gin/blob/a889c58de78711cb9b53de6cfcc9272c8518c729/recovery.go#L51
func ginRecovery(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recoverErr := recover(); recoverErr != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := recoverErr.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}
				l = l.With(
					logger.String("http_request", string(httpRequest)),
					logger.String("method", c.Request.Method),
					logger.String("url", c.Request.URL.String()),
				)
				if len(headers) > 0 {
					headersToStr := strings.Join(headers, "\r\n")
					l = l.With(logger.String("header", headersToStr))
				}

				var err error
				if handleErr, ok := recoverErr.(error); ok {
					err = errorx.Wrap(handleErr)
				} else {
					err = errorx.New("panic recovered")
				}

				if brokenPipe {
					l.WithError(err).Error("broken pipe")
				} else {
					l.WithError(err).Error("panic recovered")
				}

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					c.Error(err) //nolint: errcheck
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}
}
