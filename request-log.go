package ginmiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	bbpool "github.com/valyala/bytebufferpool"
	"github.com/waro163/gin-middleware/utils"
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
	Output        io.Writer
	Settings      []PathSettings
	MaskedFields  []string
	TimeFormat    string
	MaxBodySize   int64 // 最大body大小限制(字节)，0表示不限制
	settingsMap   map[string]*PathSettings // 内部使用的map，加速查找
}

func (l *RequestLog) initSettingsMap() {
	if l.settingsMap == nil && len(l.Settings) > 0 {
		l.settingsMap = make(map[string]*PathSettings, len(l.Settings))
		for i := range l.Settings {
			l.settingsMap[l.Settings[i].Path] = &l.Settings[i]
		}
	}
}

func (l *RequestLog) getPathSettings(path string) *PathSettings {
	if l.settingsMap != nil {
		return l.settingsMap[path]
	}
	return nil
}

// processRequestBody 根据Content-Type处理请求body，返回适合记录的JSON字节
func (l *RequestLog) processRequestBody(contentType string, body []byte) []byte {
	if len(body) == 0 {
		return []byte("{}")
	}

	// 解析Content-Type
	mediaType, _, _ := mime.ParseMediaType(contentType)

	switch mediaType {
	case "application/x-www-form-urlencoded":
		// 解析表单数据
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return body // 解析失败，返回原始数据
		}
		// 转换为map便于JSON序列化
		formData := make(map[string]interface{})
		for key, vals := range values {
			if len(vals) == 1 {
				formData[key] = vals[0]
			} else {
				formData[key] = vals
			}
		}
		jsonBytes, err := json.Marshal(formData)
		if err != nil {
			return body
		}
		return jsonBytes

	case "multipart/form-data":
		// multipart表单通常包含文件，记录一个提示信息
		return []byte(fmt.Sprintf(`{"_info":"multipart/form-data","_size":%d}`, len(body)))

	case "application/json":
		// JSON数据直接返回
		return body

	case "text/plain", "text/html", "text/xml":
		// 文本类型，包装成JSON字符串
		jsonBytes, err := json.Marshal(map[string]string{"_text": string(body)})
		if err != nil {
			return body
		}
		return jsonBytes

	default:
		// 其他类型，记录基本信息
		if strings.HasPrefix(mediaType, "text/") {
			// 其他文本类型
			jsonBytes, _ := json.Marshal(map[string]string{"_text": string(body)})
			return jsonBytes
		}
		// 二进制数据，只记录大小和类型
		return []byte(fmt.Sprintf(`{"_contentType":"%s","_size":%d}`, mediaType, len(body)))
	}
}

func (l *RequestLog) AddRequestLog() gin.HandlerFunc {
	if l.Output == nil {
		l.Output = os.Stdout
	}
	if l.TimeFormat == "" {
		l.TimeFormat = "2006-01-02 15:04:05.000"
	}
	// 初始化settings map，提高查找性能
	l.initSettingsMap()

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := CtxLogger(ctx).Output(l.Output)
		start := time.Now()
		host := c.Request.Host
		uri := c.Request.RequestURI
		method := c.Request.Method

		// 检查当前路径的配置
		pathSettings := l.getPathSettings(c.Request.URL.Path)

		// 如果配置了DisableLogs，则完全跳过日志打印和body读取
		if pathSettings != nil && pathSettings.DisableLogs {
			c.Next()
			return
		}

		// 判断是否需要读取body（优化：只在需要时读取）
		needReadReqBody := pathSettings == nil || !pathSettings.DisableBodies
		needReadRespBody := pathSettings == nil || !pathSettings.DisableBodies

		headerBytes, _ := json.Marshal(c.Request.Header)
		subLog := logger.With().
			Str(TimeStamp, start.Format(l.TimeFormat)).
			Str(Method, method).
			Str(Host, host).
			Str(Uri, uri).
			Logger()

		var reqBody []byte
		if needReadReqBody {
			// 读取请求body（带大小限制）
			if l.MaxBodySize > 0 {
				reqBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, l.MaxBodySize))
			} else {
				reqBody, _ = io.ReadAll(c.Request.Body)
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(reqBody))
		}

		// 记录请求日志
		reqLogCtx := subLog.With().RawJSON(ReqHeader, headerBytes)
		if needReadReqBody && len(reqBody) > 0 {
			// 根据Content-Type处理body数据
			contentType := c.Request.Header.Get("Content-Type")
			processedBody := l.processRequestBody(contentType, reqBody)
			reqLogCtx = reqLogCtx.RawJSON(ReqBody, processedBody)
		}
		reqLog := reqLogCtx.Logger()
		reqLog.Info().Msg(RequestLogChar)

		var bbuffer *bbpool.ByteBuffer
		// 只在需要记录响应body时才包装writer
		if needReadRespBody {
			bbuffer = bbpool.Get()
			defer bbpool.Put(bbuffer)

			// if status is 404 or 405, gin will handle those at last, but our *utils.WrapResponseWriter had been put nil
			status := c.Writer.Status()
			if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
				c.Writer = &utils.WrapResponseWriter{
					ResponseWriter: c.Writer,
					ByteBuffer:     bbuffer,
				}
			}
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		respHeaderBytes, _ := json.Marshal(c.Writer.Header())

		// 根据DisableBodies配置决定是否打印响应body数据
		respLogCtx := subLog.With().
			Int(StatusCode, status).
			RawJSON(RespHeader, respHeaderBytes)
		if needReadRespBody && bbuffer != nil && bbuffer.Len() > 0 {
			// 根据响应Content-Type处理body数据
			respContentType := c.Writer.Header().Get("Content-Type")
			processedRespBody := l.processRequestBody(respContentType, bbuffer.Bytes())
			respLogCtx = respLogCtx.RawJSON(RespBody, processedRespBody)
		}
		respLog := respLogCtx.Str(Latency, fmt.Sprintf("%v", latency)).Logger()
		respLog.Info().Msg(ResponseLogChar)

	}
}
