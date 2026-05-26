package client

import (
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

type Session struct {
	client  *Client
	storage *storage_engine.Engine
}

func NewSession(client *Client, storage *storage_engine.Engine) *Session {
	return &Session{
		client:  client,
		storage: storage,
	}
}

func (s *Session) Send(val resp.Value) error {
	return s.client.Send(val)
}

func (s *Session) RemoteAddr() net.Addr {
	return s.client.RemoteAddr()
}

func (s *Session) Storage() *storage_engine.Engine {
	return s.storage
}

func (s *Session) ReadNext() (resp.Value, error) {
	receiver := s.client.Receiver()
	val, err := resp.ReadValue(receiver)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (s *Session) LogError(errStr string) error {
	log.Println(errStr)
	return s.Send(resp.BulkStringFromString(errStr))
}
