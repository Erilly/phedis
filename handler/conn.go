package handler

import (
	"bufio"
	"fmt"
	"git.btime.cn/btime_new/phedis/config"
	"io"
	"net"
	"time"
)

type ChanBuf struct {
	Byte []byte
	Err  error
}

type Connect struct {
	Br       *bufio.Reader
	TcpConn  *net.TCPConn
	UsedAt   time.Time
	ChanRead chan *ChanBuf
	Cn       Conner
}



func ConnetRedis(opt *RedisOptions) (*Connect, error) {
	netConn, err := net.DialTimeout("tcp", opt.Addr, config.Configs.Timeout)

	if err != nil {
		return nil, err
	}

	tcpConn := netConn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)

	conn := &Connect{
		Br:       bufio.NewReader(tcpConn),
		TcpConn:  tcpConn,
		UsedAt:   time.Now(),
		ChanRead: make(chan *ChanBuf, 1),
	}
	conn.Cn = &Redis{
		conn: conn,
	}

	go conn.Cn.Read()

	err = conn.Cn.Auth(opt.Pwd)
	if err != nil {
		conn.Close()
		return nil, err
	}
	err = conn.Cn.Select(opt.Db)

	return conn, err
}

func (c *Connect) Write(b []byte) error {
	_, err := c.TcpConn.Write(b)
	c.UsedAt = time.Now()
	return err
}

func (c *Connect) Ping() error {
	return nil
}

func (c *Connect) IsActive(timeout time.Duration) bool {
	return true
}

func (c *Connect) GetReadChan() <-chan *ChanBuf {
	return c.ChanRead
}

func (c *Connect) Reply()[]byte{
	return c.Cn.Reply()
}

func (c *Connect) Close() error {
	err := c.TcpConn.Close()
	return err
}


func (c *Connect) ReadReply() ([]byte, error) {
	line, err := c.readLine()

	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, fmt.Errorf("short response line")
	}
	switch line[0] {
	case '+':
		switch string(line[1:]) {
		case "OK":
			// Avoid allocation for frequent "+OK" response.
			return line, nil
		case "PONG":
			// Avoid allocation in PING command benchmarks :)
			return line, nil
		default:
			return line, nil
		}
	case '-':
		return line, nil
	case ':':
		return line,nil
	case '$':
		n, err := parseLen(line[1:])

		if n < 0 || err != nil {
			return line, err
		}
		data := line
		p := make([]byte, n+2)
		_, err = io.ReadFull(c.Br, p)

		data = append(data,p...)

		if err != nil {
			return nil, err
		}

		//if line, err := c.readLine(); err != nil { //
		//	return nil, err
		//} else if len(line) != 0 {
		//	return nil, protocolError("bad bulk string format")
		//}

		return data, nil
	case '*':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return line, err
		}
		var r []byte
		data := line

		for i:=0;i<n;i++{
			r, err = c.ReadReply()

			data = append(data,r...)
			if err != nil {
				return nil, err
			}
		}

		return data, nil
	}
	return nil, fmt.Errorf("unexpected response line")
}

func (c *Connect) readLine() ([]byte, error) {
	// To avoid allocations, attempt to read the line using ReadSlice. This
	// call typically succeeds. The known case where the call fails is when
	// reading the output from the MONITOR command.
	p, err := c.Br.ReadSlice('\n')

	if err == bufio.ErrBufferFull {
		// The line does not fit in the bufio.Reader's buffer. Fall back to
		// allocating a buffer for the line.
		buf := append([]byte{}, p...)
		for err == bufio.ErrBufferFull {
			p, err = c.Br.ReadSlice('\n')
			buf = append(buf, p...)
		}
		p = buf
	}
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, fmt.Errorf("bad response line terminator")
	}

	return p, nil
}

func parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, fmt.Errorf("malformed length")
	}

	i := len(p) - 2
	p = p[0:i]

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return -1, fmt.Errorf("illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}