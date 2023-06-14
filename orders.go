package dvotcWS

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fasthttp/websocket"
)

type Order struct {
	QuoteID      string  `json:"quoteId,omitempty"`
	OrderType    string  `json:"orderType,omitempty"`
	Asset        string  `json:"asset,omitempty"`
	CounterAsset string  `json:"counterAsset,omitempty"`
	Price        float64 `json:"price,omitempty"`
	LimitPrice   *string `json:"limitPrice,omitempty"`
	Qty          float64 `json:"qty,omitempty"`
	Side         string  `json:"side,omitempty"`
	ClientTag    string  `json:"clientTag,omitempty"`
}

type OrderStatus struct {
	ID           string  `json:"_id"`
	ClientTag    string  `json:"clientTag"`
	LimitPrice   string  `json:"limitPrice"`
	Price        float64 `json:"price"`
	Quantity     float64 `json:"quantity"`
	Side         string  `json:"side"`
	OrderType    string  `json:"orderType,omitempty"`
	Asset        string  `json:"asset"`
	CounterAsset string  `json:"counterAsset"`
	Status       string  `json:"status"`
	User         User    `json:"user"`

	CreatedAt   time.Time  `json:"createdAt"`
	FilledAt    *time.Time `json:"filledAt"`
	CancelledAt *time.Time `json:"cancelledAt"`
}

type MarketOrderParams struct {
	QuoteID      string  `json:"quoteId"`
	Asset        string  `json:"asset"`
	CounterAsset string  `json:"counterAsset"`
	Price        float64 `json:"price"`
	Qty          float64 `json:"qty"`
	Side         string  `json:"side"`
	ClientTag    string  `json:"clientTag"`
}

type LimitOrderParams struct {
	Asset        string  `json:"asset"`
	CounterAsset string  `json:"counterAsset"`
	LimitPrice   float64 `json:"limitPrice"`
	Qty          float64 `json:"qty"`
	Side         string  `json:"side"`
	ClientTag    string  `json:"clientTag"`
}

func (dvotc *DVOTCClient) PlaceMarketOrder(marketOrder MarketOrderParams) (*OrderStatus, error) {
	conn, err := dvotc.getConn("/websocket")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	order := &Order{
		QuoteID:      marketOrder.QuoteID,
		OrderType:    "market",
		Asset:        marketOrder.Asset,
		CounterAsset: marketOrder.CounterAsset,
		Price:        marketOrder.Price,
		Qty:          marketOrder.Qty,
		Side:         marketOrder.Side,
		ClientTag:    marketOrder.ClientTag,
	}

	data, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	payload := Payload{
		Type:  MessageTypeRequestResponse,
		Event: dvotc.getRequestID(),
		Topic: "createorder",
		Data:  data,
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	resp := &Payload{}
	if err := conn.ReadJSON(resp); err != nil {
		return nil, err
	}
	if resp.Type == "error" {
		return nil, fmt.Errorf("%s", string(resp.Data))
	}
	orderRes := OrderStatus{}
	if err := json.Unmarshal(resp.Data, &orderRes); err != nil {
		return nil, err
	}

	return &orderRes, nil
}

func (dvotc *DVOTCClient) PlaceLimitOrder(limitOrder LimitOrderParams) (*OrderStatus, error) {
	conn, err := dvotc.getConn("/websocket")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sellPriceStr := fmt.Sprintf("%g", limitOrder.LimitPrice)
	order := Order{
		OrderType:    "LIMIT",
		Asset:        limitOrder.Asset,
		CounterAsset: limitOrder.CounterAsset,
		LimitPrice:   &sellPriceStr,
		Qty:          limitOrder.Qty,
		Side:         limitOrder.Side,
		ClientTag:    limitOrder.ClientTag,
	}

	data, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	payload := Payload{
		Type:  MessageTypeRequestResponse,
		Event: dvotc.getRequestID(),
		Topic: "createorder",
		Data:  data,
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	resp := &Payload{}
	if err := conn.ReadJSON(resp); err != nil {
		return nil, err
	}
	if resp.Type == "error" {
		return nil, fmt.Errorf("%s", string(resp.Data))
	}

	orderRes := OrderStatus{}
	if err := json.Unmarshal(resp.Data, &orderRes); err != nil {
		return nil, err
	}

	return &orderRes, nil
}

func (dvotc *DVOTCClient) CancelOrder(orderID string) error {
	conn, err := dvotc.getConn("/websocket")
	if err != nil {
		return err
	}
	defer conn.Close()
	payload := Payload{
		Type:  MessageTypeRequestResponse,
		Event: dvotc.getRequestID(),
		Topic: fmt.Sprintf("cancelorder/%s", orderID),
		Data:  nil,
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
		return errors.New(string(resp.Data))
	}

	return nil
}

func (dvotc *DVOTCClient) SubscribeOrderChanges(status string) (*Subscription[OrderStatus], error) {
	conn, err := dvotc.getConn("/websocket")
	if err != nil {
		return nil, err
	}
	payload := Payload{
		Type:  MessageTypeSubscribe,
		Event: "order-updates",
		Topic: fmt.Sprintf("order/%s", status),
		Data:  nil,
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	sub := &Subscription[OrderStatus]{
		Data:  make(chan OrderStatus, 100),
		done:  make(chan struct{}),
		conn:  conn,
		topic: payload.Topic,
		event: payload.Event,
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
					if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
						// server closed connection
						log.Default().Print("server closed connection")
					}
					continue
				}
				switch resp.Type {
				case MessageTypeError:
					return
				case MessageTypeInfo:
					if resp.Event == "reconnect" {
						newConn, err := dvotc.retryConnWithPayload(payload)
						if err != nil {
							// can't do much after all retries fail
							fmt.Println(err)
							return
						}
						sub.conn = newConn
					}
					continue
				}

				orderStatus := OrderStatus{}
				if err := json.Unmarshal(resp.Data, &orderStatus); err != nil {
					return
				}
				sub.Data <- orderStatus
			}
		}
	}()

	return sub, nil
}
