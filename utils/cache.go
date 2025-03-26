package utils

import (
	"fmt"
	"sync"
	"time"
)

const (
	EXPIRE_KEY = "%s:expire_at"
)

type ICache interface {
	Set(string, interface{}, time.Duration) error
	Get(string) (interface{}, error)
}

var DefaultMemoryCache ICache = &memoryCache{
	data: make(map[string]interface{}),
}

type memoryCache struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func (m *memoryCache) Set(key string, value interface{}, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
	if duration == 0 {
		return nil
	}
	expireKey := fmt.Sprintf(EXPIRE_KEY, key)
	m.data[expireKey] = time.Now().Add(duration).UnixMilli()
	return nil
}

func (m *memoryCache) Get(key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data == nil {
		return nil, nil
	}
	value, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	expireKey := fmt.Sprintf(EXPIRE_KEY, key)
	expireTime, ok := m.data[expireKey].(int64)
	if !ok {
		return value, nil
	}
	if expireTime >= time.Now().UnixMilli() {
		return value, nil
	}
	return nil, nil
}
