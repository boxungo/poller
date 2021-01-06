package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/boxungo/poller"
)

var host = flag.String("host", "localhost", "host")
var port = flag.String("port", "8089", "port")

var epoll *poller.Epoll

func main() {
	flag.Parse()

	conn, err := net.Dial("tcp", *host+":"+*port)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer conn.Close()

	epoll, err = poller.New()
	if err != nil {
		fmt.Println("Error poller New:", err)
		os.Exit(1)
	}

	fd, err := poller.SocketFD(conn)
	if err != nil {
		fmt.Println("Error poller.SocketFD", err)
		os.Exit(1)
	}

	if err := epoll.AddRead(fd); err != nil {
		fmt.Println("Error AddRead", err)
		os.Exit(1)
	}

	go func() {
		for {
			syscall.Write(fd, []byte(time.Now().Format("2006-01-02 15:04:05")))
			time.Sleep(time.Second)
		}
	}()

	err = epoll.Wait(func(ev poller.Event) error {
		if (ev.Type() & poller.REvent) != 0 {
			go read(ev)
		}

		return nil
	})

	fmt.Println("Wait error: ", err)
}

func read(ev poller.Event) {
	var buf = make([]byte, 128)
	m, err := syscall.Read(ev.FD(), buf)
	if m == 0 || err != nil {
		fmt.Println("m&err: ", m, err)
		return
	}

	fmt.Println(string(buf))

	epoll.ModRead(ev.FD())
}
