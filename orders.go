package dvotcWS

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dv-chain/dvotc-websocket-go/proto"
	"github.com/fasthttp/websocket"
	gproto "google.golang.org/protobuf/proto"
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

func (dvotc *DVOTCClient) readOrderMessageLoop(conn *websocket.Conn) {
	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				// server closed connection
				log.Print("server closed connection")
			}
			log.Print("closing connection")
			return
		}

		res := &proto.ClientMessage{}
		if err := gproto.Unmarshal(bytes, res); err != nil {
			log.Print("error decoding")
			dispatchTradeData(dvotc.orderChanStore, &dvotc.chanMutex, res.GetEvent(), res.GetTopic(), nil, err)
			continue
		}

		if errRes := res.GetErrorMessage(); errRes != nil {
			log.Println(errRes.Message)
			dispatchTradeData(dvotc.orderChanStore, &dvotc.chanMutex, res.GetEvent(), res.GetTopic(), nil, errors.New(errRes.Message))
			continue
		}

		data := res.GetTradeStatusResponse()
		if data == nil {
			log.Print("no level data ", data)
			dispatchTradeData(dvotc.orderChanStore, &dvotc.chanMutex, res.GetEvent(), res.GetTopic(), nil, errors.New("no valid response"))
			continue
		}
		dispatchTradeData(dvotc.orderChanStore, &dvotc.chanMutex, res.GetEvent(), res.GetTopic(), data.Trades[0], nil)
	}
}

func dispatchTradeData(safeChanStore map[string]tradeData, mutex *sync.RWMutex, event, topic string, trade *proto.Trade, err error) {
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

func storeTradeAndErrorChann(safeChanStore map[string]tradeData, mutex *sync.RWMutex, event, topic string, channel chan *proto.Trade, errChan chan error) {
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
func cleanupTradeAndErrorChan(safeChanStore map[string]tradeData, mutex *sync.RWMutex, event, topic string) {
	mutex.Lock()
	defer mutex.Unlock()
	key := fmt.Sprintf("%s:%s", topic, event)
	delete(safeChanStore, key)
}

type OrderResponseData = Subscription[*proto.Trade]

func (dvotc *DVOTCClient) PlaceMarketBuyOrder(marketOrder MarketOrderParams) (*proto.Trade, error) {
	conn, err := dvotc.getConnOrReuse(connectionOrders)
	if err != nil {
		return nil, err
	}

	order := &proto.CreateOrderMessage{
		QuoteId:      marketOrder.QuoteID,
		OrderType:    proto.OrderType_MARKET,
		Asset:        marketOrder.Asset,
		CounterAsset: marketOrder.CounterAsset,
		Price:        marketOrder.Price,
		Qty:          marketOrder.Qty,
		Side:         proto.OrderSide_Buy,
		ClientTag:    marketOrder.ClientTag,
		Expires:      -1,
	}
	payload := &proto.ClientMessage{
		Type:  proto.Types_requestresponse,
		Event: dvotc.getRequestID(),
		Topic: "createorder",
		Data: &proto.ClientMessage_CreateOrderRequest{
			CreateOrderRequest: order,
		},
	}

	sub := &OrderResponseData{
		Data:  make(chan *proto.Trade),
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

	data, err := gproto.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if err := dvotc.writeBinaryMessage(conn, data); err != nil {
		return nil, err
	}

	select {
	case res := <-sub.Data:
		return res, nil
	case err := <-sub.Error:
		return nil, err
	}
}

func (dvotc *DVOTCClient) PlaceMarketSellOrder(marketOrder MarketOrderParams) (*proto.Trade, error) {
	conn, err := dvotc.getConnOrReuse(connectionOrders)
	if err != nil {
		return nil, err
	}

	order := &proto.CreateOrderMessage{
		QuoteId:      marketOrder.QuoteID,
		OrderType:    proto.OrderType_MARKET,
		Asset:        marketOrder.Asset,
		CounterAsset: marketOrder.CounterAsset,
		Price:        marketOrder.Price,
		Qty:          marketOrder.Qty,
		Side:         proto.OrderSide_Sell,
		ClientTag:    marketOrder.ClientTag,
		Expires:      -1,
	}
	payload := &proto.ClientMessage{
		Type:  proto.Types_requestresponse,
		Event: dvotc.getRequestID(),
		Topic: "createorder",
		Data: &proto.ClientMessage_CreateOrderRequest{
			CreateOrderRequest: order,
		},
	}

	sub := &OrderResponseData{
		Data:  make(chan *proto.Trade),
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

	data, err := gproto.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if err := dvotc.writeBinaryMessage(conn, data); err != nil {
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
