package ginmiddleware

import (
	"context"

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

func GetContextMetaData(ctx context.Context) []ContextMetaData {
	if ctxMetas, ok := ctx.Value(ContextMetaKey(ContextKey)).([]ContextMetaData); ok {
		return ctxMetas
	}
	return make([]ContextMetaData, 0)
}

func WithMoreContextMeta(ctx context.Context, data ...ContextMetaData) context.Context {
	rawData := GetContextMetaData(ctx)
	rawData = append(rawData, data...)
	newCtx := context.WithValue(ctx, ContextMetaKey(ContextKey), rawData)
	return newCtx
}

func CtxLogger(ctx context.Context) *zerolog.Logger {
	data := GetContextMetaData(ctx)
	logContext := log.With()
	for i := range data {
		logContext = logContext.Str(data[i].Key, data[i].Value)
	}
	logger := logContext.Logger()
	return &logger
}
