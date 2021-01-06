package poller

import (
	"net"
)

// Conn 继承 net.Conn，增加文件描述符fd
type Conn struct {
	net.Conn
	fd int
}

// NewConn 新连接
func NewConn(conn net.Conn) (*Conn, error) {
	fd, err := SocketFD(conn)
	if err != nil {
		return nil, err
	}
	return &Conn{
		fd:   fd,
		Conn: conn,
	}, nil
}

// FD 返回socket文件描述符
func (c *Conn) FD() int {
	return c.fd
}
