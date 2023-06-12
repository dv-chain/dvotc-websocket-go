package dvotcWS

import (
	"github.com/fasthttp/websocket"
)

type Subscription[T any] struct {
	Data     chan T
	conn     *websocket.Conn
	done     chan struct{}
	isClosed bool
	topic    string
	event    string
	chanIdx  int

	dvotc *DVOTCClient
}

func (s *Subscription[_]) StopConsuming() error {
	// right now only levels allows broadcast
	if s.event == "levels" {
		return cleanupChannelForSymbol(&s.dvotc.safeChanStore, &s.dvotc.chanMutex, s.event, s.topic, s.chanIdx)
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
