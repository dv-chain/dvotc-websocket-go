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

type DVOTCClient struct {
	wsURL     string
	apiSecret string
	apiKey    string

	requestID int

	wsClient *websocket.Conn
	/* storing all channels to dispatch data */
	safeChanStore sync.Map
	chanMutex     sync.RWMutex

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
	// most default values are good enough
	// read more https://pkg.go.dev/github.com/avast/retry-go#pkg-variables
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
	},
		retry.Delay(1*time.Second))

	return
}

func (dvotc *DVOTCClient) getConn() (*websocket.Conn, error) {
	// need it in milliseconds
	ts := time.Now().UnixMilli()
	var timeWindow int64 = 20000

	msg := fmt.Sprintf("%s%d%d", dvotc.apiKey, ts, timeWindow)

	h := hmac.New(sha256.New, []byte(dvotc.apiSecret))
	if _, err := h.Write([]byte(msg)); err != nil {
		return nil, err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	u, err := url.Parse(dvotc.wsURL + "/websocket")
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

func (dvotc *DVOTCClient) getConnOrReuse() (*websocket.Conn, error) {
	if dvotc.wsClient != nil {
		return dvotc.wsClient, nil
	}
	// need it in milliseconds
	ts := time.Now().UnixMilli()
	var timeWindow int64 = 20000

	msg := fmt.Sprintf("%s%d%d", dvotc.apiKey, ts, timeWindow)

	h := hmac.New(sha256.New, []byte(dvotc.apiSecret))
	if _, err := h.Write([]byte(msg)); err != nil {
		return nil, err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	u, err := url.Parse(dvotc.wsURL + "/ws")
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

	dvotc.wsClient = c
	go dvotc.readMessageLoop()
	return c, nil
}

func (dvotc *DVOTCClient) readMessageLoop() {
	for {
		_, bytes, err := dvotc.wsClient.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				// server closed connection
				log.Print("server closed connection")
			}
			return
		}

		res := &proto.ClientMessage{}
		if err := gproto.Unmarshal(bytes, res); err != nil {
			log.Print("error decoding")
			return
		}

		switch res.GetType() {
		case proto.Types_error:
			return
		case proto.Types_info:
			if res.GetEvent() == "reconnect" {
				conn, err := dvotc.getConn()
				if err != nil {
					log.Println(err)
					return
				}
				reSubscribeToTopics(conn, &dvotc.safeChanStore, &dvotc.chanMutex)
				dvotc.wsClient = conn
				continue
			}
			log.Printf("invalid event type from info topic: (%s)", res.GetEvent())
		}

		levelData := res.GetLevelData()
		if levelData == nil {
			log.Print("no level data")
			return
		}
		dispatchLevelData(&dvotc.safeChanStore, &dvotc.chanMutex, res.GetEvent(), res.GetTopic(), *levelData)
	}
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

func channelsEmpty[K chan proto.LevelData](channels []K) bool {
	if len(channels) == 0 {
		return false
	}
	for _, c := range channels {
		if c != nil {
			return true
		}
	}
	return false
}

func reSubscribeToTopics(conn *websocket.Conn, mutextConn *sync.Map, mutex *sync.RWMutex) {
	mutex.Lock()
	defer mutex.Unlock()
	mutextConn.Range(func(k, value any) bool {
		channels, ok := value.([]chan proto.LevelData)
		if !ok {
			log.Fatalf("casting to type channel failed")
		}
		if channelsEmpty(channels) {
			return true
		}
		keys := strings.Split(k.(string), ":")
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

		log.Println("writing message from reconnect")
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			log.Fatal(err)
		}
		return true
	})
}

func dispatchLevelData(safeChanStore *sync.Map, mutex *sync.RWMutex, event, topic string, data proto.LevelData) {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	v, ok := safeChanStore.Load(key)
	if !ok {
		log.Fatalf("failed to dispatch level data no channel found %s", key)
	}
	channels, ok := v.([]chan proto.LevelData)
	if !ok {
		log.Fatalf("casting to type channel failed")
	}
	for _, channel := range channels {
		if channel != nil {
			channel <- data
		}
	}
}

func checkConnExistAndReturnIdx(safeChanStore *sync.Map, mutex *sync.RWMutex, event, topic string, channel chan proto.LevelData) (int, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	idx := 0
	key := fmt.Sprintf("%s:%s", topic, event)
	v, ok := safeChanStore.Load(key)
	if ok {
		channels, ok := v.([]chan proto.LevelData)
		if !ok {
			log.Fatalf("casting to type channel failed")
		}
		idx = len(channels)
		ok = channelsEmpty(channels)
		channels = append(channels, channel)
		safeChanStore.Store(key, channels)
	} else {
		safeChanStore.Store(key, []chan proto.LevelData{channel})
	}
	return idx, ok
}

func cleanupChannelForSymbol(safeChanStore *sync.Map, mutex *sync.RWMutex, event, topic string, channelIdx int) error {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	v, ok := safeChanStore.Load(key)
	if ok {
		channels, ok := v.([]chan proto.LevelData)
		if !ok {
			log.Fatalf("casting to type channel failed")
		}
		if channelIdx > len(channels)-1 {
			log.Fatalf("failed to cleanup channel not existent")
		}
		if channels[channelIdx] == nil {
			return ErrSubscriptionAlreadyClosed
		}
		close(channels[channelIdx])
		channels[channelIdx] = nil
		safeChanStore.Store(key, channels)
	}
	return nil
}
