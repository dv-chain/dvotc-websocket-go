package dvotcWS

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
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

type OrderResponseData = Subscription[*OrderStatus]

func (dvotc *DVOTCClient) PlaceMarketOrder(marketOrder MarketOrderParams) (*OrderStatus, error) {
	conn, err := dvotc.getConnOrReuse(connectionOrders)
	if err != nil {
		return nil, err
	}

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

	sub := &OrderResponseData{
		Data:  make(chan *OrderStatus),
		Error: make(chan error),
		topic: payload.Topic,
		event: payload.Event,
		conn:  conn,
	}

	storeTradeAndErrorChann(dvotc.orderChanStore, &dvotc.chanMutex, sub.event, sub.topic, sub.Data, sub.Error)
	defer func() {
		close(sub.Data)
		close(sub.Error)
		cleanupTradeAndErrorChan(dvotc.orderChanStore, &dvotc.chanMutex, sub.event, sub.topic)
	}()

	if err := dvotc.writeJSONMessage(conn, payload); err != nil {
		return nil, err
	}

	select {
	case res := <-sub.Data:
		return res, nil
	case err := <-sub.Error:
		return nil, err
	}
}

func (dvotc *DVOTCClient) PlaceLimitOrder(limitOrder LimitOrderParams) (*OrderStatus, error) {
	conn, err := dvotc.getConnOrReuse(connectionOrders)
	if err != nil {
		return nil, err
	}

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

	sub := &OrderResponseData{
		Data:  make(chan *OrderStatus),
		Error: make(chan error),
		topic: payload.Topic,
		event: payload.Event,
		conn:  conn,
	}

	storeTradeAndErrorChann(dvotc.orderChanStore, &dvotc.chanMutex, sub.event, sub.topic, sub.Data, sub.Error)
	defer func() {
		close(sub.Data)
		close(sub.Error)
		cleanupTradeAndErrorChan(dvotc.orderChanStore, &dvotc.chanMutex, sub.event, sub.topic)
	}()

	if err := dvotc.writeJSONMessage(conn, payload); err != nil {
		return nil, err
	}

	select {
	case res := <-sub.Data:
		return res, nil
	case err := <-sub.Error:
		return nil, err
	}
}

func (dvotc *DVOTCClient) CancelOrder(orderID string) error {
	conn, err := dvotc.getConn()
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
	conn, err := dvotc.getConn()
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

func (dvotc *DVOTCClient) readOrderMessageLoop(conn *websocket.Conn, cleanupFunc func()) {
	for {
		resp := Payload{}
		if err := conn.ReadJSON(&resp); err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				// server closed connection
				log.Default().Print("server closed connection")
			}
			if cleanupFunc != nil {
				cleanupFunc()
			}
			return
		}

		// handle error
		if resp.Type == "error" {
			errResp := ErrorResponse{}
			if err := json.Unmarshal(resp.Data, &errResp); err != nil {
				dispatchTradeData(dvotc.orderChanStore, &dvotc.chanMutex, resp.Event, resp.Topic, nil, err)
				continue
			}
			dispatchTradeData(dvotc.orderChanStore, &dvotc.chanMutex, resp.Event, resp.Topic, nil, errors.New(errResp.Message))
			continue
		}

		orderStatus := OrderStatus{}
		if err := json.Unmarshal(resp.Data, &orderStatus); err != nil {
			return
		}
		dispatchTradeData(dvotc.orderChanStore, &dvotc.chanMutex, resp.Event, resp.Topic, &orderStatus, nil)
	}
}

func dispatchTradeData(safeChanStore map[string]tradeData, mutex *sync.RWMutex, event, topic string, trade *OrderStatus, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	channels, ok := safeChanStore[key]
	if !ok {
		log.Panicf("key already exist for trade %s", key)
	}
	if err != nil {
		channels.err <- err
	} else {
		channels.data <- trade
	}
}

func storeTradeAndErrorChann(safeChanStore map[string]tradeData, mutex *sync.RWMutex, event, topic string, channel chan *OrderStatus, errChan chan error) {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	_, ok := safeChanStore[key]
	if ok {
		log.Panicf("key already exist for trade %s", key)
	}

	safeChanStore[key] = tradeData{
		data: channel,
		err:  errChan,
	}
}

// cleanupTradeAndErrorChan remove the tradeData channels for a key
// TODO: make this cleanup on a timer basis
func cleanupTradeAndErrorChan(safeChanStore map[string]tradeData, mutex *sync.RWMutex, event, topic string) {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	delete(safeChanStore, key)
}
