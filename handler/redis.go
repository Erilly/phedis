package handler

import (
	"fmt"
	"strings"
	"time"
)


type Conner interface {
	Ping()error
	Auth(string)error
	Select(string)error
	Read()
	Reply()[]byte
}

type RedisOptions struct {
	Addr string
	Pwd  string
	Db   string
}

type Redis struct {
	conn *Connect
}

var _ Conner = (*Redis)(nil)


func (r Redis) Ping() error {
	err := r.conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		return nil
	}
	cb := <-r.conn.GetReadChan()
	if cb.Err != nil {
		return cb.Err
	}
	if strings.ToUpper(string(cb.Byte)) != "+PONG\r\n" {
		return fmt.Errorf("error ping")
	}
	return nil
}

func (r Redis) Auth(pwd string) error {
	if pwd == "" {
		return nil
	}

	data := fmt.Sprintf("*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(pwd), pwd)
	err := r.conn.Write([]byte(data))
	if err != nil {
		return err
	}
	cb := <-r.conn.GetReadChan()
	if cb.Err != nil {
		return cb.Err
	}
	if strings.ToUpper(string(cb.Byte)) != "+OK\r\n" {
		return fmt.Errorf("auth error")
	}
	return nil
}

func (r Redis) Select(db string) error {
	if db == "" {
		db = "0"
	}

	data := fmt.Sprintf("*2\r\n$6\r\nSELECT\r\n$%d\r\n%s\r\n", len(db), db)
	err := r.conn.Write([]byte(data))
	if err != nil {
		return err
	}
	cb := <-r.conn.GetReadChan()
	if cb.Err != nil {
		return cb.Err
	}
	if strings.ToUpper(string(cb.Byte)) != "+OK\r\n" {
		return fmt.Errorf("select error")
	}
	return nil
}

func (r Redis) Reply() []byte{
	cb := <-r.conn.GetReadChan()
	if cb.Err != nil {
		return []byte("")
	}
	return cb.Byte
}

func (r Redis) Read() {
	var (
		line []byte
		err  error
		cb   *ChanBuf
	)

	for {
		line, err = r.conn.ReadReply()

		r.conn.UsedAt = time.Now()
		cb = &ChanBuf{Byte: line, Err: err}
		r.conn.ChanRead <- cb

		if err != nil {
			break
		}
	}
}