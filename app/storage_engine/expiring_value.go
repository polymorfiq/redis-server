package storage_engine

import (
	"errors"
	"io"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type ExpiringValue struct {
	val       resp.Value
	expiresAt time.Time
}

func NewExpiringValue(val resp.Value, expiresAt time.Time) resp.Value {
	return &ExpiringValue{
		val:       val,
		expiresAt: expiresAt,
	}
}

func (v *ExpiringValue) Read(_ io.Reader) error {
	return errors.New("reading expiring values from a Reader is not currently implemented.")
}

func (v *ExpiringValue) WriteTo(w io.Writer) (n int64, err error) {
	return v.Value().WriteTo(w)
}

func (v *ExpiringValue) Value() resp.Value {
	if v.expiresAt.Before(time.Now()) {
		return resp.NewNull()
	}

	return v.val
}
