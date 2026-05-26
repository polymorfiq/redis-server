package client

import (
	"io"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Client struct {
	conn net.Conn
}

func New(conn net.Conn) *Client {
	return &Client{conn: conn}
}

func (c *Client) Receiver() io.Reader {
	return c.conn
}

func (c *Client) Send(val resp.Value) error {
	_, err := val.WriteTo(c.conn)
	return err
}

func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
