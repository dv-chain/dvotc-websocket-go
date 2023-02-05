package dvotcWS

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fasthttp/websocket"
)

type LoginNotification struct {
	IP        string           `json:"ip"`
	User      NotificationUser `json:"user"`
	UserName  string           `json:"userName"`
	UserEmail string           `json:"userEmail"`
	Text      string           `json:"text"`
}

type NotificationUser struct {
	UserID       int64           `json:"id"`
	ID           string          `json:"_id"`
	FirstName    string          `json:"firstName"`
	LastName     string          `json:"lastName"`
	CompanyName  string          `json:"companyName"`
	UUID         string          `json:"uuid"`
	GroupAccount json.RawMessage `json:"groupAccount,omitempty"`
}

type BatchCreatedNotification struct {
	BatchDetails []BatchDetail    `json:"batchDetails"`
	UserName     string           `json:"userName"`
	UserEmail    string           `json:"userEmail"`
	BatchUUID    string           `json:"batch_uuid"`
	User         NotificationUser `json:"user"`
	Text         string           `json:"text"`
}

type BatchDetail struct {
	Symbol      string `json:"symbol"`
	NetQuantity string `json:"net_quantity"`
}

type OrderNotification struct {
	Asset        string           `json:"asset"`
	CounterAsset string           `json:"counterAsset"`
	ID           string           `json:"_id"`
	Quantity     int64            `json:"quantity"`
	Price        float64          `json:"price"`
	LimitPrice   float64          `json:"limitPrice"`
	Side         string           `json:"side"`
	OrderType    string           `json:"orderType"`
	Source       string           `json:"source"`
	Status       string           `json:"status"`
	FilledAt     *time.Time       `json:"filledAt"`
	CreatedAt    time.Time        `json:"createdAt"`
	CancelledAt  *time.Time       `json:"cancelledAt"`
	ClientTag    string           `json:"clientTag"`
	User         NotificationUser `json:"user"`
	UserName     string           `json:"userName"`
	UserEmail    string           `json:"userEmail"`
	Text         string           `json:"text"`
}

type LimitReached80Notification struct {
	Asset     string           `json:"asset"`
	Unsettled float64          `json:"unsettled"`
	BuyMax    int64            `json:"buyMax"`
	SellMax   int64            `json:"sellMax"`
	Message   string           `json:"message"`
	User      NotificationUser `json:"user"`
	UserName  string           `json:"userName"`
	UserEmail string           `json:"userEmail"`
	Text      string           `json:"text"`
}

const (
	NOTIFICAITON_LOGIN            = "LOGIN"
	NOTIFICATION_BATCH_CREATED    = "BATCH_CREATED"
	NOTIFICATION_BATCH_SETTLED    = "BATCH_SETTLED" // implement me
	NOTIFICATION_ORDER_CREATED    = "ORDER_CREATED"
	NOTIFICATION_ORDER_FILLED     = "ORDER_FILLED"
	NOTIFICATION_ORDER_CANCELLED  = "ORDER_CANCELLED"
	NOTIFICATION_SETTLEMENT_ADDED = "SETTLEMENT_ADDED" // implement me

	NOTIFICATION_LIMIT_CHANGED    = "LIMIT_CHANGED"    // implement me
	NOTIFICATION_LIMIT_REACHED_80 = "LIMIT_REACHED_80" // implement me
	NOTIFICATION_LIMIT_REACHED_95 = "LIMIT_REACHED_95" // implement me

)

// LIMIT_CHANGED, LIMIT_REACHED_80, IMIT_REACHED_95, BATCH_CREATED, SETTLEMENT_ADDED, BATCH_SETTLED, LOGIN, ORDER_CREATED, ORDER_CANCELLED, ORDER_FILLED
func (dvotc *DVOTCClient) SubscribeLogin() (*Subscription[LoginNotification], error) {
	sub, err := SubscribeNotifications[LoginNotification](dvotc, NOTIFICAITON_LOGIN)
	return sub, err
}

func (dvotc *DVOTCClient) SubscribeBatchCreated() (*Subscription[BatchCreatedNotification], error) {
	sub, err := SubscribeNotifications[BatchCreatedNotification](dvotc, NOTIFICATION_BATCH_CREATED)
	return sub, err
}

// func (dvotc *DVOTCClient) SubscribeOrderCreated() (*Subscription[OrderNotification], error) {
// 	sub, err := SubscribeNotifications[OrderNotification](dvotc, NOTIFICATION_ORDER_CREATED)
// 	return sub, err
// }

// func (dvotc *DVOTCClient) SubscribeOrderFilled() (*Subscription[OrderNotification], error) {
// 	sub, err := SubscribeNotifications[OrderNotification](dvotc, NOTIFICATION_ORDER_FILLED)
// 	return sub, err
// }

// func (dvotc *DVOTCClient) SubscribeOrderCancelled() (*Subscription[OrderNotification], error) {
// 	sub, err := SubscribeNotifications[OrderNotification](dvotc, NOTIFICATION_ORDER_CANCELLED)
// 	return sub, err
// }

// func (dvotc *DVOTCClient) SubscribeLimitReached80() (*Subscription[LimitReached80Notification], error) {
// 	sub, err := SubscribeNotifications[LimitReached80Notification](dvotc, NOTIFICATION_LIMIT_REACHED_80)
// 	return sub, err
// }

func SubscribeNotifications[
	K LoginNotification |
		BatchCreatedNotification |
		OrderNotification |
		LimitReached80Notification,
](dvotc *DVOTCClient, topic string) (*Subscription[K], error) {
	conn, err := dvotc.getConn()
	if err != nil {
		return nil, err
	}
	payload := Payload{
		Type:  MessageTypeSubscribe,
		Event: "notifications",
		Topic: topic,
		Data:  nil,
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	sub := &Subscription[K]{
		Data:  make(chan K, 100),
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
					return
				}
				fmt.Println(resp)
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
				case MessageTypePingPong:
					continue
				}

				var payload K
				if err := json.Unmarshal(resp.Data, &payload); err != nil {
					return
				}
				sub.Data <- payload
			}
		}
	}()

	// keep connection alive
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
