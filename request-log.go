package ginmiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestLogChar  string = "RequestLog"
	ResponseLogChar string = "ResponseLog"
	TimeStamp       string = "TimeStamp"
	Method          string = "Method"
	Host            string = "Host"
	Uri             string = "Uri"
	ReqHeader       string = "ReqHeader"
	ReqBody         string = "ReqBody"
	RespHeader      string = "RespHeader"
	RespBody        string = "RespBody"
	StatusCode      string = "StatusCode"
	Latency         string = "Latency"
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
	TimeFormat   string
}

func (l *RequestLog) AddRequestLog() gin.HandlerFunc {
	if l.Output == nil {
		l.Output = os.Stdout
	}
	if l.TimeFormat == "" {
		l.TimeFormat = "2006-01-02 15:04:05.000"
	}
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := CtxLogger(ctx).Output(l.Output)
		start := time.Now()
		host := c.Request.Host
		uri := c.Request.RequestURI
		method := c.Request.Method
		headerBytes, _ := json.Marshal(c.Request.Header)
		subLog := logger.With().
			Str(TimeStamp, start.Format(l.TimeFormat)).
			Str(Method, method).
			Str(Host, host).
			Str(Uri, uri).
			Logger()

		reqBody, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(reqBody))

		reqLog := subLog.With().
			RawJSON(ReqHeader, headerBytes).
			RawJSON(ReqBody, reqBody).
			Logger()
		reqLog.Info().Msg(RequestLogChar)

		c.Next()

		latency := time.Since(start)
		respHeaderBytes, _ := json.Marshal(c.Writer.Header())
		respLog := subLog.With().
			Int(StatusCode, c.Writer.Status()).
			RawJSON(RespHeader, respHeaderBytes).
			RawJSON(RespBody, []byte("TBD")).
			Str(Latency, fmt.Sprintf("%v", latency)).
			Logger()
		respLog.Info().Msg(ResponseLogChar)

	}
}
