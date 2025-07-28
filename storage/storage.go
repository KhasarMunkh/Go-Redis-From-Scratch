package storage

import (
	"sync"
)

type Getter interface {
	Get(key string) (val string, ok bool)
}

type Setter interface {
	Set(key, value string)
}

type Deleter interface {
	Delete(keys ...string) int
}

type Pinger interface {
	Ping()
}

type Storage interface {
	Getter
	Setter 
	Deleter
	// will add more methods like TTL, EXPIRE, etc.
}

type entry struct {
	Value string
	// will add more fields like expiration time, etc.
}

type MemoryStorage struct {
	data map[string]entry
	mu   sync.RWMutex
	// will add channel for expiration, etc.
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]entry),
	}
}

func (m *MemoryStorage) Get(key string) (val string, ok bool) {
	m.mu.RLock()
	v, ok := m.data[key]
	m.mu.RUnlock()
	return v.Value, ok
}

func (m *MemoryStorage) Set(key string, val string) {
	m.mu.Lock()
	m.data[key] = entry{Value: val}
	m.mu.Unlock()
}

func (m *MemoryStorage) Delete(keys ...string) (n int) {
	m.mu.Lock()
	for _, k := range keys {
		if _, ok := m.data[k]; ok {
			delete(m.data, k)
			n++
		}
	}
	m.mu.Unlock()
	return n
}
