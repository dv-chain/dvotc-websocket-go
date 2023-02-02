package dvotcWS

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/fasthttp/websocket"
)

var (
	ErrClientConnectionNotFound  = errors.New("client connection not established")
	ErrInvalidPayload            = errors.New("invalid paload returned")
	ErrSubscriptionAlreadyClosed = errors.New("subscription is already closed")
)

type MessageType string

const (
	MessageTypeSubscribe       MessageType = "subscribe"
	MessageTypeUnsubscribe     MessageType = "unsubscribe"
	MessageTypeRequestResponse MessageType = "request-response"
	MessageTypePingPong        MessageType = "ping-pong"
	MessageTypeInfo            MessageType = "info"
	MessageTypeError           MessageType = "error"
)

type DVOTCClient struct {
	wsURL     string
	apiSecret string
	apiKey    string

	requestID int64

	mu sync.Mutex
}

type Payload struct {
	Type  MessageType     `json:"type"`
	Topic string          `json:"topic"`
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

func NewDVOTCClient(wsURL, apiKey, apiSecret string) *DVOTCClient {
	return &DVOTCClient{
		wsURL:     wsURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		requestID: 10,
	}
}

func (dvotc *DVOTCClient) retryConnWithPayload(payload Payload) (conn *websocket.Conn, err error) {
	// default values are good enough https://pkg.go.dev/github.com/avast/retry-go#pkg-variables
	err = retry.Do(func() error {
		conn, err = dvotc.getConn()
		if err != nil {
			return err
		}

		err = conn.WriteJSON(payload)
		if err != nil {
			return err
		}
		return nil
	})

	return
}

func (dvotc *DVOTCClient) getConn() (*websocket.Conn, error) {
	// need it in milliseconds
	ts := time.Now().UnixNano() / int64(time.Millisecond)
	var timeWindow int64 = 20000

	msg := fmt.Sprintf("%s%d%d", dvotc.apiKey, ts, timeWindow)

	h := hmac.New(sha256.New, []byte(dvotc.apiSecret))
	if _, err := h.Write([]byte(msg)); err != nil {
		return nil, err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	u, err := url.Parse(dvotc.wsURL)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	header.Set("dv-timestamp", fmt.Sprintf("%d", ts))
	header.Set("dv-timewindow", fmt.Sprintf("%d", timeWindow))
	header.Set("dv-signature", signature)
	header.Set("dv-api-key", dvotc.apiKey)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (dvotc *DVOTCClient) getRequestID() string {
	dvotc.mu.Lock()
	defer dvotc.mu.Unlock()
	reqID := dvotc.requestID
	dvotc.requestID += 1
	return fmt.Sprintf("%d", reqID)
}

func (dvotc *DVOTCClient) Ping() error {
	conn, err := dvotc.getConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	payload := &Payload{
		Type:  MessageTypePingPong,
		Event: dvotc.getRequestID(),
		Topic: "Topiccs",
	}
	err = conn.WriteJSON(payload)
	if err != nil {
		return err
	}

	resp := &Payload{}
	if err := conn.ReadJSON(resp); err != nil {
		return err
	}
	if resp.Type == "error" {
		return fmt.Errorf("returned error with message: %s", string(resp.Data))
	}
	return nil
}
