package dvotcWS

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/dv-chain/dvotc-websocket-go/proto"
	"github.com/fasthttp/websocket"
	gproto "google.golang.org/protobuf/proto"
)

var (
	ErrClientConnectionNotFound  = errors.New("client connection not established")
	ErrInvalidPayload            = errors.New("invalid payload returned")
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

type tradeData struct {
	data chan *proto.Trade
	err  chan error
}

type DVOTCClient struct {
	wsURL     string
	apiSecret string
	apiKey    string

	requestID int

	wsConnStore map[connectionTypes]*websocket.Conn
	/* storing all channels to dispatch data */
	levelChanStore map[string][]chan *proto.LevelData
	orderChanStore map[string]tradeData

	chanMutex sync.RWMutex
	mu        sync.Mutex
}

type Payload struct {
	Type  MessageType     `json:"type"`
	Topic string          `json:"topic"`
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type connectionTypes int64

const (
	connectionNew connectionTypes = iota
	connectionLevel
	connectionOrders
)

func NewDVOTCClient(wsURL, apiKey, apiSecret string) *DVOTCClient {
	return &DVOTCClient{
		wsURL:          wsURL,
		apiKey:         apiKey,
		apiSecret:      apiSecret,
		wsConnStore:    make(map[connectionTypes]*websocket.Conn),
		orderChanStore: make(map[string]tradeData),
		levelChanStore: make(map[string][]chan *proto.LevelData),
		requestID:      10,
	}
}

func (dvotc *DVOTCClient) retryConnWithPayload(payload Payload) (conn *websocket.Conn, err error) {
	// most default values are good enough
	// read more https://pkg.go.dev/github.com/avast/retry-go#pkg-variables
	err = retry.Do(func() error {
		conn, err = dvotc.getConn("/websocket")
		if err != nil {
			return err
		}

		err = conn.WriteJSON(payload)
		if err != nil {
			return err
		}
		return nil
	},
		retry.Delay(1*time.Second))

	return
}

func (dvotc *DVOTCClient) getConn(path string) (*websocket.Conn, error) {
	// need it in milliseconds
	ts := time.Now().UnixMilli()
	var timeWindow int64 = 20000

	msg := fmt.Sprintf("%s%d%d", dvotc.apiKey, ts, timeWindow)

	h := hmac.New(sha256.New, []byte(dvotc.apiSecret))
	if _, err := h.Write([]byte(msg)); err != nil {
		return nil, err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	u, err := url.Parse(dvotc.wsURL + path)
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

func (dvotc *DVOTCClient) getConnOrReuse(t connectionTypes) (*websocket.Conn, error) {
	dvotc.mu.Lock()
	defer dvotc.mu.Unlock()
	conn, ok := dvotc.wsConnStore[t]
	if ok {
		return conn, nil
	}

	c, err := dvotc.getConn("/ws")
	if err != nil {
		return nil, err
	}
	dvotc.wsConnStore[t] = c

	switch t {
	case connectionLevel:
		go dvotc.readLevelMessageLoop(c)
	case connectionOrders:
		go dvotc.readOrderMessageLoop(c)
	}
	return c, nil
}

// writeBinaryMessage allows to write only one message to connection
func (dvotc *DVOTCClient) writeBinaryMessage(conn *websocket.Conn, b []byte) error {
	dvotc.mu.Lock()
	defer dvotc.mu.Unlock()
	return conn.WriteMessage(websocket.BinaryMessage, b)
}

func (dvotc *DVOTCClient) getRequestID() string {
	dvotc.mu.Lock()
	defer dvotc.mu.Unlock()
	reqID := dvotc.requestID
	dvotc.requestID += 1
	return fmt.Sprintf("%d", reqID)
}

func (dvotc *DVOTCClient) Ping() error {
	conn, err := dvotc.getConn("/websocket")
	if err != nil {
		return err
	}
	defer conn.Close()

	payload := &Payload{
		Type:  MessageTypePingPong,
		Event: dvotc.getRequestID(),
		Topic: "ping-pong",
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

func reSubscribeToTopics(conn *websocket.Conn, levelChanStore map[string][]chan *proto.LevelData, mutex *sync.RWMutex) {
	mutex.Lock()
	defer mutex.Unlock()
	for k, v := range levelChanStore {
		if channelsEmpty(v) {
			continue
		}
		keys := strings.Split(k, ":")
		topic, event := keys[0], keys[1]
		payload := &proto.ClientMessage{
			Type:  proto.Types_subscribe,
			Event: event,
			Topic: topic,
		}

		data, err := gproto.Marshal(payload)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("writing message from reconnect ", payload)
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			log.Fatal(err)
		}
	}
}
