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
}

func (s *Subscription[_]) StopConsuming() error {
	if s.isClosed {
		return ErrSubscriptionAlreadyClosed
	}
	s.isClosed = true
	close(s.done)
	<-s.Data

	payload := Payload{
		Type:  MessageTypeUnsubscribe,
		Event: s.event,
		Topic: s.topic,
	}
	err := s.conn.WriteJSON(payload)
	if err != nil {
		return err
	}

	return s.conn.Close()
}
