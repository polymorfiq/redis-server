package storage_engine

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Engine struct {
	data map[string]resp.Value
}

func New() *Engine {
	return &Engine{
		data: make(map[string]resp.Value),
	}
}

func (e *Engine) Put(key string, value resp.Value, opts StorageOpts) error {
	storedVal := value
	if opts.expires {
		storedVal = NewExpiringValue(storedVal, opts.expiresAt)
	}

	e.data[key] = storedVal
	return nil
}

func (e *Engine) Get(key string) (resp.Value, bool) {
	val, ok := e.data[key]
	return val, ok
}

type StorageOpts struct {
	expires   bool
	expiresAt time.Time
}

func (opts *StorageOpts) SetExpiresAt(expiresAt time.Time) {
	opts.expiresAt = expiresAt
	opts.expires = true
}
