package poller

import (
	"syscall"
)

// Epoll ...
type Epoll struct {
	epfd int
}

// New ...
func New() (*Epoll, error) {
	l := new(Epoll)
	epfd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}
	l.epfd = epfd
	return l, nil
}

// Close 关闭连接
func (p *Epoll) Close() error {
	return syscall.Close(p.epfd)
}

// Wait 等待事件
func (p *Epoll) Wait(iter func(ev Event) error) error {
	events := make([]syscall.Kevent_t, 128)
	for {
		n, err := syscall.Kevent(p.epfd, nil, events, nil)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return err
		}

		for i := 0; i < n; i++ {
			if events[i].Ident != 0 {
				ev := Event{fd: int(events[i].Ident), etype: p.event(events[i].Filter)}
				if err := iter(ev); err != nil {
					return err
				}
			}
		}
	}
}

// Add ...
func (p *Epoll) Add(fd int, changes []syscall.Kevent_t) error {
	_, err := syscall.Kevent(p.epfd, changes, nil, nil)
	return err
}

// Mod ...
func (p *Epoll) Mod(fd int, changes []syscall.Kevent_t) error {
	_, err := syscall.Kevent(p.epfd, changes, nil, nil)
	return err
}

// Del ...
func (p *Epoll) Del(fd int) error {
	changes := []syscall.Kevent_t{
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_DELETE | syscall.EV_ONESHOT, Filter: syscall.EVFILT_WRITE,
		},
	}
	_, err := syscall.Kevent(fd, changes, nil, nil)

	return err
}

// AddRead ...
func (p *Epoll) AddRead(fd int) error {
	changes := []syscall.Kevent_t{
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_ONESHOT, Filter: syscall.EVFILT_READ,
		},
	}
	return p.Add(fd, changes)
}

// AddWrite ...
func (p *Epoll) AddWrite(fd int) error {
	changes := []syscall.Kevent_t{
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_ONESHOT, Filter: syscall.EVFILT_WRITE,
		},
	}
	return p.Add(fd, changes)
}

// AddReadWrite ...
func (p *Epoll) AddReadWrite(fd int) error {
	changes := []syscall.Kevent_t{
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_ONESHOT, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_ONESHOT, Filter: syscall.EVFILT_WRITE,
		},
	}
	return p.Add(fd, changes)
}

// ModRead ...
func (p *Epoll) ModRead(fd int) error {
	changes := []syscall.Kevent_t{
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_ONESHOT, Filter: syscall.EVFILT_READ,
		},
	}
	return p.Mod(fd, changes)
}

// ModWrite ...
func (p *Epoll) ModWrite(fd int) error {
	changes := []syscall.Kevent_t{
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_ONESHOT, Filter: syscall.EVFILT_WRITE,
		},
	}
	return p.Mod(fd, changes)
}

// ModReadWrite ...
func (p *Epoll) ModReadWrite(fd int) error {
	return p.AddReadWrite(fd)
}

// Event
func (p *Epoll) event(filter int16) EventType {
	switch filter {
	case syscall.EVFILT_WRITE:
		return WEvent
	case syscall.EVFILT_READ:
		return REvent
	}
	return REvent
}
