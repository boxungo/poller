package poller

import (
	"net"
)

// Listener only TCP mode links are supported
type Listener struct {
	net.Listener
	fd int
}

// NewListener ...
func NewListener(ln net.Listener) (Listener, error) {
	fd, err := SocketFD(ln)
	if err != nil {
		return Listener{}, err
	}
	return Listener{
		Listener: ln,
		fd:       fd,
	}, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (ln *Listener) Close() error {
	return ln.Listener.Close()
}

// FD return the listener fd
func (ln *Listener) FD() int {
	return ln.fd
}

// Accept ...
func (ln *Listener) Accept() (*Conn, error) {
	conn, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewConn(conn)
}
