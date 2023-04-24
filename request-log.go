package ginmiddleware

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestLogChar string = "RequestLog"
	TimeStamp      string = "TimeStamp"
	Method         string = "Method"
	Host           string = "Host"
	Uri            string = "Uri"
	ReqHeader      string = "ReqHeader"
	ReqBody        string = "ReqBody"
	RespHeader     string = "RespHeader"
	RespBody       string = "RespBody"
	StatusCode     string = "StatusCode"
	Latency        string = "Latency"
)

type PathSettings struct {
	Path          string
	DisableLogs   bool
	DisableBodies bool
}

type RequestLog struct {
	Output       io.Writer
	Settings     []PathSettings
	MaskedFields []string
}

func (l *RequestLog) AddRequestLog() gin.HandlerFunc {
	if l.Output == nil {
		l.Output = os.Stdout
	}
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := CtxLogger(ctx).Output(l.Output)
		start := time.Now()
		host := c.Request.Host
		uri := c.Request.RequestURI
		method := c.Request.Method
		headerBytes, _ := json.Marshal(c.Request.Header)

		c.Next()

		latency := time.Since(start)
		respHeaderBytes, _ := json.Marshal(c.Writer.Header())
		subLog := logger.With().Str(TimeStamp, start.String()).
			Str(Method, method).
			Str(Host, host).
			Str(Uri, uri).
			RawJSON(ReqHeader, headerBytes).
			Int(StatusCode, c.Writer.Status()).
			RawJSON(RespHeader, respHeaderBytes).
			Str(Latency, fmt.Sprintf("%v", latency)).
			Logger()
		subLog.Info().Msg(RequestLogChar)

	}
}
