package ginmiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var HeaderXRequestID string = "X-Request-ID"

func AddRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqID string
		if requestIDInHeader := c.GetHeader(HeaderXRequestID); requestIDInHeader != "" {
			reqID = requestIDInHeader
		} else {
			reqID = uuid.New().String()
			c.Request.Header.Set(HeaderXRequestID, reqID)
		}

		c.Set(HeaderXRequestID, reqID)
		// set request id on request's context
		c.Request = c.Request.WithContext(
			WithMoreContextMeta(c.Request.Context(), ContextMetaData{Key: HeaderXRequestID, Value: reqID}),
		)

		c.Next()
	}
}
