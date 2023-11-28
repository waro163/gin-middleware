package ginmiddleware

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	ContextKey = "contextMetaKey"
)

type ContextMetaKey string

type ContextMetaData struct {
	Key   string
	Value string
}

// GetContextMetaData get GetContextMetaData slice from incoming context
func GetContextMetaData(ctx context.Context) []ContextMetaData {
	if ctxMetas, ok := ctx.Value(ContextMetaKey(ContextKey)).([]ContextMetaData); ok {
		return ctxMetas
	}
	return make([]ContextMetaData, 0)
}

// WithMoreContextMeta add extra ContextMetaData to incoming context
func WithMoreContextMeta(ctx context.Context, data ...ContextMetaData) context.Context {
	rawData := GetContextMetaData(ctx)
	rawData = append(rawData, data...)
	newCtx := context.WithValue(ctx, ContextMetaKey(ContextKey), rawData)
	return newCtx
}

// CtxLogger put the ContextMetaData of incoming context into log
func CtxLogger(ctx context.Context) *zerolog.Logger {
	data := GetContextMetaData(ctx)
	logContext := log.With()
	for i := range data {
		logContext = logContext.Str(data[i].Key, data[i].Value)
	}
	logger := logContext.Logger()
	return &logger
}

func InjectContextRequestHeader(ctx context.Context, req *http.Request) {
	ctxMetas := GetContextMetaData(ctx)
	for i := range ctxMetas {
		ctxMeta := ctxMetas[i]
		req.Header.Add(ctxMeta.Key, ctxMeta.Value)
	}
}

// NewContextWithMeta new a context with incoming context metadata
// when we want to run a go routine in background, we should use this function
// and also can get those metadata from output context
func NewContextWithMeta(ctx context.Context) context.Context {
	rawData := GetContextMetaData(ctx)
	return context.WithValue(context.Background(), ContextMetaKey(ContextKey), rawData)
}
