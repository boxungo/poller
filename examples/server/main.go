package main

import (
	"fmt"
	"net"
	"syscall"

	"github.com/boxungo/poller"
)

type Server struct {
	ln    poller.Listener
	epoll *poller.Epoll
	fds   map[int]int
}

func (s *Server) Serve() {
	listener, err := net.Listen("tcp", ":8089")
	if err != nil {
		fmt.Println("net.Listen error: ", err)
		return
	}

	ln, err := poller.NewListener(listener)
	if err != nil {
		fmt.Println("poller.NewListener error: ", err)
		return
	}
	defer ln.Close()
	s.ln = ln

	fmt.Println("Listener FD: ", s.ln.FD())

	epoll, err := poller.New()
	if err != nil {
		fmt.Println("poller.New error: ", err)
		return
	}
	defer epoll.Close()
	s.epoll = epoll

	epoll.AddRead(s.ln.FD())
	err = s.Wait()

	if err != nil {
		fmt.Println("epoll.Wait error:", err)
		return
	}
}

func (s *Server) Wait() error {
	return s.epoll.Wait(func(ev poller.Event) error {
		if s.ln.FD() == ev.FD() {
			go s.Accept(ev)
			return nil
		}

		if (ev.Type() & poller.REvent) != 0 {
			go s.Read(ev)
		}

		return nil
	})
}

func (s *Server) Accept(ev poller.Event) {
	conn, err := s.ln.Accept()
	if err != nil {
		fmt.Println("Accept error: ", err)
		return
	}
	fmt.Println("Accept")

	if err := s.epoll.ModRead(ev.FD()); err != nil {
		fmt.Println("Accept s.epoll.ModRead(ev.FD()) error: ", err)
		return
	}
	s.fds[conn.FD()] = conn.FD()

	if err := s.epoll.AddRead(conn.FD()); err != nil {
		fmt.Println("Accept s.epoll.AddRead(conn.FD()) error: ", err)
		return
	}

}

func (s *Server) Read(ev poller.Event) {
	var buf = make([]byte, 128)
	m, err := syscall.Read(ev.FD(), buf)
	if m == 0 || err != nil {
		fmt.Println("m&err: ", m, err)
		return
	}

	str := string(buf[:m])
	if str == "QUIT" {
		fmt.Println(ev.FD(), " QUIT")
		syscall.Close(ev.FD())
		delete(s.fds, ev.FD())
		return
	}

	fmt.Println("buf: ", str)

	if err := s.epoll.ModRead(ev.FD()); err != nil {
		fmt.Println("Read ModRead(ev.FD()) error: ", err)
		return
	}

	// todo write
	for _, fd := range s.fds {
		syscall.Write(fd, []byte(fmt.Sprintf("I receive %s\r\n", str)))
	}

	return
}

func main() {
	srv := &Server{
		fds: make(map[int]int, 32),
	}

	srv.Serve()

}
