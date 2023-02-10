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
	UserID       int64            `json:"id"`
	ID           string           `json:"_id"`
	FirstName    string           `json:"firstName"`
	LastName     string           `json:"lastName"`
	CompanyName  string           `json:"companyName"`
	UUID         string           `json:"uuid"`
	GroupAccount *json.RawMessage `json:"groupAccount,omitempty"`
}

type BatchNotificaiton struct {
	Info         *Info            `json:"info,omitempty"`
	BatchDetails []BatchDetail    `json:"batchDetails"`
	UserName     string           `json:"userName"`
	UserEmail    string           `json:"userEmail"`
	BatchUUID    string           `json:"batch_uuid"`
	User         NotificationUser `json:"user"`
	Text         string           `json:"text"`
}
type BatchCreatedNotification = BatchNotificaiton

type BatchSettledNotification = BatchNotificaiton

type SettlementAddedNotification = BatchSettledNotification

type Info struct {
	ID          int64       `json:"id"`
	Type        string      `json:"type"`
	BatchID     int64       `json:"batch_id"`
	UserID      int64       `json:"user_id"`
	AssetID     int64       `json:"asset_id"`
	NetQuantity json.Number `json:"net_quantity" faker:"oneof: -5, 6"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	UpdatedBy   int64       `json:"updated_by"`
	Reference   string      `json:"reference"`
	Asset       string      `json:"asset"`
}

type BatchDetail struct {
	Symbol      string      `json:"symbol"`
	NetQuantity json.Number `json:"net_quantity" faker:"oneof: -5, 6"`
}

type OrderNotification struct {
	Asset        string           `json:"asset"`
	CounterAsset string           `json:"counterAsset"`
	ID           string           `json:"_id"`
	Quantity     int64            `json:"quantity"`
	Price        float64          `json:"price"`
	LimitPrice   json.Number      `json:"limitPrice"`
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

// type LimitReached80Notification struct {
// 	Asset     string           `json:"asset"`
// 	Unsettled float64          `json:"unsettled"`
// 	BuyMax    int64            `json:"buyMax"`
// 	SellMax   int64            `json:"sellMax"`
// 	Message   string           `json:"message"`
// 	User      NotificationUser `json:"user"`
// 	UserName  string           `json:"userName"`
// 	UserEmail string           `json:"userEmail"`
// 	Text      string           `json:"text"`
// }

type LimitChangedNotification struct {
	Changes   []Change         `json:"changes"`
	User      NotificationUser `json:"user"`
	UserName  string           `json:"userName"`
	UserEmail string           `json:"userEmail"`
	Text      string           `json:"text"`
}

type Change struct {
	Symbol           string      `json:"symbol"`
	OldMin           json.Number `json:"oldMin" faker:"oneof: -5, 6"`
	NewMin           json.Number `json:"newMin" faker:"oneof: -5, 6"`
	OldMax           json.Number `json:"oldMax" faker:"oneof: -5, 6"`
	NewMax           json.Number `json:"newMax" faker:"oneof: -5, 6"`
	OldUnsettledBuy  json.Number `json:"oldUnsettledBuy" faker:"oneof: -5, 6"`
	NewUnsettledBuy  json.Number `json:"newUnsettledBuy" faker:"oneof: -5, 6"`
	OldUnsettledSell json.Number `json:"oldUnsettledSell" faker:"oneof: -5, 6"`
	NewUnsettledSell json.Number `json:"newUnsettledSell" faker:"oneof: -5, 6"`
}

const (
	NOTIFICAITON_LOGIN            = "LOGIN"
	NOTIFICATION_BATCH_CREATED    = "BATCH_CREATED"
	NOTIFICATION_BATCH_SETTLED    = "BATCH_SETTLED"
	NOTIFICATION_SETTLEMENT_ADDED = "SETTLEMENT_ADDED"
	NOTIFICATION_ORDER_CREATED    = "ORDER_CREATED"
	NOTIFICATION_ORDER_FILLED     = "ORDER_FILLED"
	NOTIFICATION_ORDER_CANCELLED  = "ORDER_CANCELLED"

	NOTIFICATION_LIMIT_CHANGED = "LIMIT_CHANGED"
	// NOTIFICATION_LIMIT_REACHED_80 = "LIMIT_REACHED_80" // disabled
	// NOTIFICATION_LIMIT_REACHED_95 = "LIMIT_REACHED_95" // disabled

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

func (dvotc *DVOTCClient) SubscribeBatchSettled() (*Subscription[BatchSettledNotification], error) {
	sub, err := SubscribeNotifications[BatchSettledNotification](dvotc, NOTIFICATION_BATCH_SETTLED)
	return sub, err
}

func (dvotc *DVOTCClient) SubscribeSettlementAdded() (*Subscription[SettlementAddedNotification], error) {
	sub, err := SubscribeNotifications[SettlementAddedNotification](dvotc, NOTIFICATION_SETTLEMENT_ADDED)
	return sub, err
}

func (dvotc *DVOTCClient) SubscribeOrderCreated() (*Subscription[OrderNotification], error) {
	sub, err := SubscribeNotifications[OrderNotification](dvotc, NOTIFICATION_ORDER_CREATED)
	return sub, err
}

func (dvotc *DVOTCClient) SubscribeOrderFilled() (*Subscription[OrderNotification], error) {
	sub, err := SubscribeNotifications[OrderNotification](dvotc, NOTIFICATION_ORDER_FILLED)
	return sub, err
}

func (dvotc *DVOTCClient) SubscribeOrderCancelled() (*Subscription[OrderNotification], error) {
	sub, err := SubscribeNotifications[OrderNotification](dvotc, NOTIFICATION_ORDER_CANCELLED)
	return sub, err
}

func (dvotc *DVOTCClient) SubscribeLimitChanged() (*Subscription[LimitChangedNotification], error) {
	sub, err := SubscribeNotifications[LimitChangedNotification](dvotc, NOTIFICATION_LIMIT_CHANGED)
	return sub, err
}

func SubscribeNotifications[
	K LoginNotification |
		BatchNotificaiton |
		OrderNotification |
		LimitChangedNotification,
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
				if resp.Topic != "ping-pong" {
					fmt.Println(resp.Event, resp.Topic, resp.Type, string(resp.Data))
				}
				switch resp.Type {
				case MessageTypeError:
					return
				case MessageTypeInfo:
					if resp.Event == "reconnect" {
						sub.conn = nil
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
					fmt.Println(err)
					return
				}
				sub.Data <- payload
			}
		}
	}()

	// keep connection alive
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-sub.done:
				ticker.Stop()
				return
			case <-ticker.C:
				if sub.conn == nil {
					// means its reconnecting
					continue
				}
				payload := &Payload{
					Type:  MessageTypePingPong,
					Event: dvotc.getRequestID(),
					Topic: "ping-pong",
				}
				if err := sub.conn.WriteJSON(payload); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}()

	return sub, nil
}
