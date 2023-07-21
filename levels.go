package dvotcWS

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/fasthttp/websocket"
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

type SubscribeLevelData = Subscription[LevelData]

type LevelSubscription struct {
	Data    *FIFOQueue[LevelData]
	topic   string
	event   string
	chanIdx int
}

func (dvotc *DVOTCClient) SubscribeLevels(symbol string) (*LevelSubscription, error) {
	sub := &LevelSubscription{
		Data:    NewFIFOQueue[LevelData](),
		topic:   symbol,
		event:   "levels",
		chanIdx: 0,
	}

	chanIdx, ok := checkLevelsConnExistAndReturnIdx(dvotc.levelChanStore, &dvotc.chanMutex, sub.event, sub.topic, sub.Data)
	sub.chanIdx = chanIdx
	if ok {
		// just add a new channel to list to listen to subscriptions
		return sub, nil
	}

	conn, err := dvotc.getConnOrReuse(connectionLevel)
	if err != nil {
		return nil, err
	}
	payload := Payload{
		Type:  MessageTypeSubscribe,
		Event: "levels",
		Topic: symbol,
	}

	if err := dvotc.writeJSONMessage(conn, payload); err != nil {
		return nil, err
	}

	return sub, nil
}

func (dvotc *DVOTCClient) readLevelMessageLoop(conn *websocket.Conn) {
	var err error
	for {
		resp := Payload{}
		if err := conn.ReadJSON(&resp); err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				// server closed connection
				log.Default().Print("server closed connection")
			}
			return
		}
		switch resp.Type {
		case MessageTypeError:
			return
		case MessageTypeInfo:
			if resp.Event == "reconnect" {
				conn, err = dvotc.getConn()
				if err != nil {
					log.Println(err)
					return
				}
				reSubscribeToTopics(conn, dvotc.levelChanStore, &dvotc.chanMutex)
			}
			continue
		}

		levelData := &LevelData{}
		if err := json.Unmarshal(resp.Data, levelData); err != nil {
			log.Println(err)
			return
		}
		dispatchLevelData(dvotc.levelChanStore, &dvotc.chanMutex, resp.Event, resp.Topic, levelData)
	}
}

func channelsEmpty(channels []*FIFOQueue[LevelData]) bool {
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

func dispatchLevelData(safeChanStore map[string][]*FIFOQueue[LevelData], mutex *sync.RWMutex, event, topic string, data *LevelData) {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	channels, ok := safeChanStore[key]
	if !ok {
		log.Fatalf("casting to type channel failed")
	}
	for _, channel := range channels {
		if channel != nil {
			channel.enqueue(data)
		}
	}
}

func checkLevelsConnExistAndReturnIdx(safeChanStore map[string][]*FIFOQueue[LevelData], mutex *sync.RWMutex, event, topic string, channel *FIFOQueue[LevelData]) (int, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	idx := 0
	key := fmt.Sprintf("%s:%s", topic, event)

	channels, existingConnection := safeChanStore[key]
	if existingConnection {
		idx = len(channels)
		existingConnection = !channelsEmpty(channels)
		channels = append(channels, channel)
		safeChanStore[key] = channels
	} else {
		safeChanStore[key] = []*FIFOQueue[LevelData]{channel}
	}
	return idx, existingConnection
}
