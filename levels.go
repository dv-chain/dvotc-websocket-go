package dvotcWS

import (
	"encoding/json"
	"fmt"
	"log"

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
				if err := sub.conn.ReadJSON(&resp); err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
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
						newConn, err := dvotc.retryConnWithPayload(payload)
						if err != nil {
							fmt.Println(err)
							// can't do much after all retries fail
							return
						}
						sub.conn = newConn
						fmt.Println("after assignment")
					}
					continue
				case MessageTypePingPong:
					continue
				}
				levelData := LevelData{}
				if err := json.Unmarshal(resp.Data, &levelData); err != nil {
					return
				}
				sub.Data <- levelData
			}
		}
	}()

	return sub, nil
}
