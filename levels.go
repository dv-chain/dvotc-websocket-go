package dvotcWS

import (
	"fmt"
	"log"
	"sync"

	"github.com/dv-chain/dvotc-websocket-go/proto"
	"github.com/fasthttp/websocket"
	gproto "google.golang.org/protobuf/proto"
)

type LevelData struct {
	Levels     []Level `json:"levels"`
	LastUpdate int64   `json:"lastUpdate"`
	QuoteID    string  `json:"quoteId"`
	Market     string  `json:"market"`
}

type Level struct {
	SellPrice   float64 `json:"sellPrice"`
	BuyPrice    float64 `json:"buyPrice"`
	MaxQuantity float64 `json:"maxQuantity"`
}

type SubscribeLevelData = Subscription[*proto.LevelData]

func (dvotc *DVOTCClient) SubscribeLevels(symbol string) (*SubscribeLevelData, error) {
	conn, err := dvotc.getConnOrReuse(connectionLevel)
	if err != nil {
		return nil, err
	}

	sub := &SubscribeLevelData{
		Data:  make(chan *proto.LevelData, 5),
		done:  make(chan struct{}),
		topic: symbol,
		event: "levels",
		idx:   0,
		dvotc: dvotc,
	}

	chanIdx, connExist := checkLevelsConnExistAndReturnIdx(dvotc.levelChanStore, &dvotc.chanMutex, sub.event, sub.topic, sub.Data)
	sub.idx = chanIdx
	if connExist {
		return sub, nil
	}

	payload := &proto.ClientMessage{
		Type:  proto.Types_subscribe,
		Event: "levels",
		Topic: symbol,
	}

	data, err := gproto.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if err := dvotc.writeBinaryMessage(conn, data); err != nil {
		return nil, err
	}

	return sub, nil
}

func (dvotc *DVOTCClient) readLevelMessageLoop(conn *websocket.Conn) {
	for {
		_, bytes, err := conn.ReadMessage()
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
				conn, err = dvotc.getConn("/ws")
				if err != nil {
					log.Println(err)
					return
				}
				log.Printf("got reconnect message")
				reSubscribeToTopics(conn, dvotc.levelChanStore, &dvotc.chanMutex)
				continue
			}
			log.Printf("invalid event type from info topic: (%s)", res.GetEvent())
		}

		levelData := res.GetLevelData()
		if levelData == nil {
			log.Print("no level data")
			return
		}
		dispatchLevelData(dvotc.levelChanStore, &dvotc.chanMutex, res.GetEvent(), res.GetTopic(), levelData)
	}
}

func cleanupLevelChannelForSymbol(levelChanStore map[string][]chan *proto.LevelData, mutex *sync.RWMutex, event, topic string, channelIdx int) error {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	channels, ok := levelChanStore[key]
	if !ok {
		return nil
	}
	if channelIdx > len(channels)-1 {
		log.Fatalf("failed to cleanup channel not existent")
		return nil
	}
	if channels[channelIdx] == nil {
		return ErrSubscriptionAlreadyClosed
	}
	close(channels[channelIdx])
	channels[channelIdx] = nil
	levelChanStore[key] = channels
	return nil
}

func checkLevelsConnExistAndReturnIdx(levelChanStore map[string][]chan *proto.LevelData, mutex *sync.RWMutex, event, topic string, channel chan *proto.LevelData) (int, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	idx := 0
	key := fmt.Sprintf("%s:%s", topic, event)

	channels, existingConnection := levelChanStore[key]
	if existingConnection {
		idx = len(channels)
		existingConnection = !channelsEmpty(channels)
		channels = append(channels, channel)
		levelChanStore[key] = channels
	} else {
		levelChanStore[key] = []chan *proto.LevelData{channel}
	}
	return idx, existingConnection
}

func dispatchLevelData(levelChanStore map[string][]chan *proto.LevelData, mutex *sync.RWMutex, event, topic string, data *proto.LevelData) {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	channels, ok := levelChanStore[key]
	if !ok {
		log.Fatalf("casting to type channel failed")
	}
	for _, channel := range channels {
		if channel != nil {
			select {
			case channel <- data:
				// successfully sent data in channel
			default:
				// message could not be sent due to full channel
			}
		}
	}
}

func channelsEmpty(channels []chan *proto.LevelData) bool {
	if len(channels) == 0 {
		return true
	}
	for _, c := range channels {
		if c != nil {
			return false
		}
	}
	return true
}
