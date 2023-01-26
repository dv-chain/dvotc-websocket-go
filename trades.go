package dvotcWS

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Trade struct {
	ID           string     `json:"_id"`
	Price        float64    `json:"price"`
	LimitPrice   float64    `json:"limitPrice"`
	Quantity     int64      `json:"quantity"`
	Side         string     `json:"side"`
	ClientTag    string     `json:"clientTag"`
	Asset        string     `json:"asset"`
	CounterAsset string     `json:"counterAsset"`
	Status       string     `json:"status"`
	User         User       `json:"user"`
	Batch        Batch      `json:"batch"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
	FilledAt     time.Time  `json:"filledAt"`
}

type Batch struct {
	ID      string `json:"_id"`
	Settled bool   `json:"settled"`
}
type User struct {
	ID        string `json:"_id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type ListTradesPayload struct {
	IDs        string `json:"ids"`
	TradeKeys  string `json:"tradeKeys"`
	ClientTags string `json:"clientTags"`
}

func (dvotc *DVOTCClient) ListTrades(IDs []string, tradeKeys []string, clientTags []string) ([]Trade, error) {
	conn, err := dvotc.getConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data := ListTradesPayload{
		IDs:        strings.Join(IDs, ","),
		TradeKeys:  strings.Join(tradeKeys, ","),
		ClientTags: strings.Join(clientTags, ","),
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	payload := Payload{
		Type:  MessageTypeRequestResponse,
		Event: dvotc.getRequestID(),
		Topic: "tradestatus",
		Data:  dataBytes,
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	resp := &Payload{}
	if err := conn.ReadJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Type == "error" {
		return nil, fmt.Errorf("error with message: %s", string(resp.Data))
	}

	trades := make([]Trade, 0)
	if err := json.Unmarshal(resp.Data, &trades); err != nil {
		return nil, err
	}

	return trades, nil
}
