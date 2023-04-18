package ginmiddleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContextMetaData(t *testing.T) {
	data := GetContextMetaData(context.TODO())
	assert.Equal(t, len(data), 0, "should be 0")
	newCtx := WithMoreContextMeta(context.TODO(), ContextMetaData{Key: "key", Value: "value"})
	newData := GetContextMetaData(newCtx)
	assert.Equal(t, len(newData), 1, "should be 1")
}
