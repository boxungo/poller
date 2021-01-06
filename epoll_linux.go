package poller

import (
	"syscall"
)

const readEvent = syscall.EPOLLIN | syscall.EPOLLONESHOT
const writeEvent = syscall.EPOLLOUT | syscall.EPOLLONESHOT
const readWriteEvent = syscall.EPOLLIN | syscall.EPOLLOUT | syscall.EPOLLONESHOT

type Epoll struct {
	epfd   int // epoll fd
	wfd    int // event fd
	closed chan struct{}
}

func New() (*Epoll, error) {
	epoll := new(Epoll)
	// 创建Epoll文件描述符
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	epoll.epfd = epfd

	r0, _, errno := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if errno != 0 {
		syscall.Close(epoll.epfd)
		return nil, errno
	}
	epoll.wfd = int(r0)

	err = syscall.EpollCtl(epoll.epfd, syscall.EPOLL_CTL_ADD, epoll.wfd, &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(epoll.wfd),
	})
	if err != nil {
		syscall.Close(epoll.wfd)
		syscall.Close(epoll.epfd)
		return nil, err
	}
	return epoll, nil
}

func (epoll *Epoll) Close() error {
	if _, err := syscall.Write(epoll.wfd, []byte{0, 0, 0, 0, 0, 0, 0, 1}); err != nil {
		return err
	}

	<-epoll.closed

	syscall.Close(epoll.wfd)
	syscall.Close(epoll.epfd)

	return nil
}

func (epoll *Epoll) Wait(iter func(ev Event) error) error {
	defer func() {
		close(epoll.closed)
	}()

	events := make([]syscall.EpollEvent, 512)
	for {
		// epollwait 这个系统调用是用来返回epfd中的就绪的事件。events指向调用者可以使用的事件的内存区域。msec 指定超时时间
		n, err := syscall.EpollWait(epoll.epfd, events, -1)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return err
		}

		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)
			if fd == epoll.wfd {
				return nil
			}

			ev := Event{etype: epoll.event(events[i].Events), fd: fd}
			if err := iter(ev); err != nil {
				return err
			}
		}
	}

	return nil
}

func (epoll *Epoll) Add(fd int, events uint32) error {
	// EPOLL_CTL_ADD: 在epfd中注册指定的fd文件描述符并能把event和fd关联起来
	return syscall.EpollCtl(epoll.epfd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

func (epoll *Epoll) Mod(fd int, events uint32) error {
	// EPOLL_CTL_MOD: 改变 fd和event之间的联系
	return syscall.EpollCtl(epoll.epfd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

func (epoll *Epoll) Del(fd int) error {
	// EPOLL_CTL_DEL: 从指定的epfd中删除fd文件描述符。在这种模式中event是被忽略的，并且值可以等于nil
	return syscall.EpollCtl(epoll.epfd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (epoll *Epoll) AddRead(fd int) error {
	return epoll.Add(fd, readEvent)
}

func (epoll *Epoll) AddWrite(fd int) error {
	return epoll.Add(fd, writeEvent)
}

func (epoll *Epoll) AddReadWrite(fd int) error {
	return epoll.Add(fd, readWriteEvent)
}

func (epoll *Epoll) ModRead(fd int) error {
	return epoll.Mod(fd, readEvent)
}

func (epoll *Epoll) ModWrite(fd int) error {
	return epoll.Mod(fd, writeEvent)
}

func (epoll *Epoll) ModReadWrite(fd int) error {
	return epoll.Mod(fd, readWriteEvent)
}

func (epoll *Epoll) event(et uint32) EventType {
	var etype EventType
	if (et & syscall.EPOLLOUT) != 0 {
		etype |= WEvent
	}

	if (et & syscall.EPOLLIN) != 0 {
		etype |= REvent
	}

	if etype == 0 {
		etype = REvent
	}
	return etype
}
