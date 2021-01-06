package poller

// EventType 事件类型
type EventType int

// EEvent
const (
	EEvent EventType = iota << 1 // 错误
	REvent                       // 读
	WEvent                       // 写
)

// Event Epoll事件
type Event struct {
	etype EventType
	fd    int
}

// Type 事件类型
func (ev Event) Type() EventType {
	return ev.etype
}

// FD 事件描述符
func (ev Event) FD() int {
	return ev.fd
}
