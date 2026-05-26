package client

import (
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/storage_engine"
)

type Session struct {
	protoVersion int
	client       *Client
	storage      *storage_engine.Engine
}

func NewSession(client *Client, storage *storage_engine.Engine) *Session {
	return &Session{
		protoVersion: 2,
		client:       client,
		storage:      storage,
	}
}

func (s *Session) Send(val resp.Value) error {
	return s.client.Send(val)
}

func (s *Session) SetProtoVersion(version int) {
	s.protoVersion = version
}

func (s *Session) IsRESP3() bool {
	return s.protoVersion >= 3
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
	return s.Send(resp.ErrorFromString(fmt.Sprintf("ERR %s", errStr)))
}
