package utils

import (
	"sync"

	"github.com/gin-gonic/gin"
)

type WrapResponseWriter struct {
	gin.ResponseWriter
	copyData []byte
}

func (w *WrapResponseWriter) Write(data []byte) (int, error) {
	if len(w.copyData) == 0 {
		w.copyData = make([]byte, len(data))
		copy(w.copyData, data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *WrapResponseWriter) GetBody() []byte {
	return w.copyData
}

var defaultPool = sync.Pool{
	New: func() any {
		return new(WrapResponseWriter)
	},
}

func Get(w gin.ResponseWriter) *WrapResponseWriter {
	wrw := defaultPool.Get().(*WrapResponseWriter)
	wrw.ResponseWriter = w
	if wrw.copyData == nil {
		wrw.copyData = make([]byte, 0)
	}
	return wrw
}

func Put(w *WrapResponseWriter) {
	w.ResponseWriter = nil
	w.copyData = w.copyData[:0]
	defaultPool.Put(w)
}
