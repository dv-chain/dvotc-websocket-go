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
	"time"

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

	header *DVOTCHeader
}

type DVOTCHeader struct {
	Timestamp  string `json:"dv-timestamp"`
	TimeWindow string `json:"dv-timewindow"`
	Signature  string `json:"dv-signature"`
	ApiKey     string `json:"dv-api-key"`
}

type Payload struct {
	Type  MessageType     `json:"type"`
	Topic string          `json:"topic"`
	Data  json.RawMessage `json:"data"`
	Event string          `json:"event"`
}

func NewDVOTCClient(wsURL, apiKey, apiSecret string) *DVOTCClient {
	return &DVOTCClient{
		wsURL:     wsURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		requestID: 10,
	}
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

	u := url.URL{Scheme: "wss", Host: dvotc.wsURL, Path: "/websocket"}
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

func (dvotc *DVOTCClient) ListAvailableSymbols() ([]string, error) {
	conn, err := dvotc.getConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	payload := Payload{
		Type:  MessageTypeRequestResponse,
		Event: dvotc.getRequestID(),
		Topic: "availablesymbols",
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	resp := Payload{}
	err = conn.ReadJSON(&resp)
	if err != nil {
		return nil, err
	}

	var symbols []string
	err = json.Unmarshal(resp.Data, &symbols)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}
