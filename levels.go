package dvotcWS

import (
	"log"

	"github.com/dv-chain/dvotc-websocket-go/proto"
	"github.com/gorilla/websocket"
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

func (dvotc *DVOTCClient) SubscribeLevels(symbol string) (*Subscription[proto.LevelData], error) {
	sub := &Subscription[proto.LevelData]{
		Data:    make(chan proto.LevelData),
		done:    make(chan struct{}),
		topic:   symbol,
		event:   "levels",
		chanIdx: 0,
		dvotc:   dvotc,
	}

	if chanIdx, ok := checkConnExistAndReturnIdx(&dvotc.safeChanStore, sub.event, sub.topic, sub.Data); ok {
		// just add a new channel to list to listen to subscriptions
		sub.chanIdx = chanIdx
		return sub, nil
	}

	conn, err := dvotc.getConnOrReuse()
	if err != nil {
		return nil, err
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

	log.Println("writing message")
	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return nil, err
	}

	return sub, nil
}
