package dvotcWS

import (
	"github.com/fasthttp/websocket"
)

type Subscription[T any] struct {
	Data     chan T
	Error    chan error
	conn     *websocket.Conn
	done     chan struct{}
	isClosed bool
	topic    string
	event    string

	idx   int
	dvotc *DVOTCClient
}

func (s *Subscription[_]) StopConsuming() error {
	if s.event == "levels" {
		return cleanupLevelChannelForSymbol(s.dvotc.levelChanStore, &s.dvotc.chanMutex, s.event, s.topic, s.idx)
	}
	if s.isClosed {
		return ErrSubscriptionAlreadyClosed
	}
	s.isClosed = true
	close(s.done)
	<-s.Data

	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
