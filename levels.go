package dvotcWS

import (
	"encoding/json"
	"time"
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

func (dvotc *DVOTCClient) SubscribeLevels(symbol string) (*Subscription[LevelData], error) {
	conn, err := dvotc.getConn()
	if err != nil {
		return nil, err
	}
	payload := Payload{
		Type:  MessageTypeSubscribe,
		Event: "levels",
		Topic: symbol,
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	sub := &Subscription[LevelData]{
		Data:  make(chan LevelData),
		done:  make(chan struct{}),
		conn:  conn,
		topic: symbol,
		event: "level",
	}

	go func() {
		defer close(sub.Data)
		for {
			select {
			case <-sub.done:
				return
			default:
				resp := Payload{}
				if err := conn.ReadJSON(&resp); err != nil {
					continue
				}
				if resp.Type == "error" {
					return
				}
				levelData := LevelData{}
				if err := json.Unmarshal(resp.Data, &levelData); err != nil {
					return
				}
				sub.Data <- levelData
			}
		}
	}()

	// keep conneciton alive
	go func() {
		for {
			time.Sleep(time.Second * 5)
			payload := &Payload{
				Type:  MessageTypePingPong,
				Event: dvotc.getRequestID(),
				Topic: "ping-pong",
			}
			if err := conn.WriteJSON(payload); err != nil {
				return
			}
		}
	}()

	return sub, nil
}
