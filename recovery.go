package ginmiddleware

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DebugStack string = "DebugStack"
)

func Recovery(writer io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := CtxLogger(ctx).Output(writer)

		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				stack := debug.Stack()
				if e, ok := err.(error); ok {
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if current[0] == "Authorization" {
							headers[idx] = current[0] + ": *"
						}
					}
					if brokenPipe {
						logger.Err(e).Bytes(DebugStack, stack).Strs("panicHeader", headers).Msg("net op error panic")
					} else {
						logger.Err(e).Bytes(DebugStack, stack).Msg("panic error")
					}
				} else {
					logger.Error().Bytes(DebugStack, stack).Msg("panic")
				}
				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"message": "server panic"})
				}
			}
		}()
		c.Next()

	}
}
