package utils

import (
	"github.com/gin-gonic/gin"
	bbpool "github.com/valyala/bytebufferpool"
)

type WrapResponseWriter struct {
	gin.ResponseWriter
	ByteBuffer *bbpool.ByteBuffer
}

func (w *WrapResponseWriter) Write(data []byte) (int, error) {
	w.ByteBuffer.Write(data)
	return w.ResponseWriter.Write(data)
}
