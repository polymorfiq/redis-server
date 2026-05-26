package storage_engine

import (
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Engine struct {
	data           map[string]resp.Value
	changeChannels map[string]chan bool
	channelsLock   sync.Mutex
}

func New() *Engine {
	return &Engine{
		data:           make(map[string]resp.Value),
		changeChannels: make(map[string]chan bool),
	}
}

func (e *Engine) Put(key string, value resp.Value, opts StorageOpts) error {
	storedVal := value
	if opts.expires {
		storedVal = NewExpiringValue(storedVal, opts.expiresAt)
	}

	changeChannel, changeChannelExists := e.changeChannels[key]
	if changeChannelExists {
		select {
		case changeChannel <- true:
		default:
		}
	}

	e.data[key] = storedVal
	return nil
}

func (e *Engine) Get(key string) (resp.Value, bool) {
	val, ok := e.data[key]
	if !ok {
		return nil, false
	}

	expVal, mayExpire := val.(*ExpiringValue)
	if mayExpire {
		return expVal.Value()
	}

	return val, true
}

func (e *Engine) ChangeChannel(key string) <-chan bool {
	e.channelsLock.Lock()
	defer e.channelsLock.Unlock()

	changeChannel, channelExists := e.changeChannels[key]
	if !channelExists {
		e.changeChannels[key] = make(chan bool)
		return e.changeChannels[key]
	} else {
		return changeChannel
	}
}

type StorageOpts struct {
	expires   bool
	expiresAt time.Time
}

func (opts *StorageOpts) SetExpiresAt(expiresAt time.Time) {
	opts.expiresAt = expiresAt
	opts.expires = true
}
