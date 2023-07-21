package dvotcWS_test

// import (
// 	"encoding/json"
// 	"fmt"
// 	"testing"
// 	"time"

// 	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
// 	"github.com/go-faker/faker/v4"
// 	"github.com/go-faker/faker/v4/pkg/options"
// 	"github.com/stretchr/testify/require"
// )

// func TestPlaceMarketOrder(t *testing.T) {
// 	t.Run("Successfully placed order", func(t *testing.T) {
// 		p := dvotcWS.Payload{
// 			Type:  "request-response",
// 			Event: "10",
// 			Topic: "createorder",
// 		}

// 		var marketOrder dvotcWS.MarketOrderParams
// 		err := faker.FakeData(&marketOrder)
// 		require.NoError(t, err)

// 		order := &dvotcWS.Order{
// 			QuoteID:      marketOrder.QuoteID,
// 			OrderType:    "market",
// 			Asset:        marketOrder.Asset,
// 			CounterAsset: marketOrder.CounterAsset,
// 			Price:        marketOrder.Price,
// 			Qty:          marketOrder.Qty,
// 			Side:         marketOrder.Side,
// 			ClientTag:    marketOrder.ClientTag,
// 		}

// 		dataBytes, err := json.Marshal(order)
// 		require.NoError(t, err)
// 		p.Data = dataBytes
// 		reqBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		orderStatus := dvotcWS.OrderStatus{}
// 		err = faker.FakeData(&orderStatus)
// 		require.NoError(t, err)
// 		now := time.Now().UTC()
// 		orderStatus.FilledAt = &now
// 		orderStatus.CreatedAt = now
// 		orderStatus.CancelledAt = nil
// 		dataBytes, err = json.Marshal(orderStatus)
// 		require.NoError(t, err)
// 		p.Data = dataBytes
// 		respBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		wsServer := &echoWebsocketServer{
// 			t:        t,
// 			response: [][]byte{respBytes},
// 			request:  reqBytes,
// 		}

// 		url := setupTestWebsocketServer(wsServer)

// 		client := dvotcWS.NewDVOTCClient(url, "123", "321")
// 		resp, err := client.PlaceMarketOrder(marketOrder)
// 		require.NoError(t, err)

// 		fmt.Println(resp.CancelledAt, orderStatus.CancelledAt)
// 		require.Equal(t, orderStatus, *resp)
// 	})

// 	t.Run("Error returned from placing order", func(t *testing.T) {
// 		p := dvotcWS.Payload{
// 			Type:  "request-response",
// 			Event: "10",
// 			Topic: "createorder",
// 		}

// 		var marketOrder dvotcWS.MarketOrderParams
// 		err := faker.FakeData(&marketOrder)
// 		require.NoError(t, err)

// 		order := &dvotcWS.Order{
// 			QuoteID:      marketOrder.QuoteID,
// 			OrderType:    "market",
// 			Asset:        marketOrder.Asset,
// 			CounterAsset: marketOrder.CounterAsset,
// 			Price:        marketOrder.Price,
// 			Qty:          marketOrder.Qty,
// 			Side:         marketOrder.Side,
// 			ClientTag:    marketOrder.ClientTag,
// 		}

// 		dataBytes, err := json.Marshal(order)
// 		require.NoError(t, err)
// 		p.Data = dataBytes
// 		reqBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		wsServer := &echoWebsocketServer{
// 			t:        t,
// 			response: [][]byte{[]byte(`{"type": "error", "topic": "availablesymbols", "event": "10", "data": {"message": "internal server error"}}`)},
// 			request:  reqBytes,
// 		}

// 		url := setupTestWebsocketServer(wsServer)

// 		client := dvotcWS.NewDVOTCClient(url, "123", "321")
// 		resp, err := client.PlaceMarketOrder(marketOrder)
// 		require.Error(t, err)
// 		require.ErrorContains(t, err, "internal server error")

// 		require.Nil(t, resp)
// 	})
// }

// func TestPlaceLimitOrder(t *testing.T) {
// 	t.Run("Successfully placed order", func(t *testing.T) {
// 		p := dvotcWS.Payload{
// 			Type:  "request-response",
// 			Event: "10",
// 			Topic: "createorder",
// 		}

// 		var limitOrder dvotcWS.LimitOrderParams
// 		err := faker.FakeData(&limitOrder)
// 		require.NoError(t, err)

// 		sellPriceStr := fmt.Sprintf("%g", limitOrder.LimitPrice)
// 		order := dvotcWS.Order{
// 			OrderType:    "LIMIT",
// 			Asset:        limitOrder.Asset,
// 			CounterAsset: limitOrder.CounterAsset,
// 			LimitPrice:   &sellPriceStr,
// 			Qty:          limitOrder.Qty,
// 			Side:         limitOrder.Side,
// 			ClientTag:    limitOrder.ClientTag,
// 		}

// 		dataBytes, err := json.Marshal(order)
// 		require.NoError(t, err)
// 		p.Data = dataBytes
// 		reqBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		orderStatus := dvotcWS.OrderStatus{}
// 		err = faker.FakeData(&orderStatus)
// 		require.NoError(t, err)
// 		orderStatus.FilledAt = nil
// 		orderStatus.CreatedAt = time.Now().UTC()
// 		orderStatus.CancelledAt = nil
// 		dataBytes, err = json.Marshal(orderStatus)
// 		require.NoError(t, err)
// 		p.Data = dataBytes
// 		respBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		wsServer := &echoWebsocketServer{
// 			t:        t,
// 			response: [][]byte{respBytes},
// 			request:  reqBytes,
// 		}

// 		url := setupTestWebsocketServer(wsServer)

// 		client := dvotcWS.NewDVOTCClient(url, "123", "321")
// 		resp, err := client.PlaceLimitOrder(limitOrder)
// 		require.NoError(t, err)

// 		require.Equal(t, orderStatus, *resp)
// 	})

// 	t.Run("Error returned from placing order", func(t *testing.T) {
// 		p := dvotcWS.Payload{
// 			Type:  "request-response",
// 			Event: "10",
// 			Topic: "createorder",
// 		}

// 		var limitOrder dvotcWS.LimitOrderParams
// 		err := faker.FakeData(&limitOrder)
// 		require.NoError(t, err)

// 		sellPriceStr := fmt.Sprintf("%g", limitOrder.LimitPrice)
// 		order := dvotcWS.Order{
// 			OrderType:    "LIMIT",
// 			Asset:        limitOrder.Asset,
// 			CounterAsset: limitOrder.CounterAsset,
// 			LimitPrice:   &sellPriceStr,
// 			Qty:          limitOrder.Qty,
// 			Side:         limitOrder.Side,
// 			ClientTag:    limitOrder.ClientTag,
// 		}

// 		dataBytes, err := json.Marshal(order)
// 		require.NoError(t, err)
// 		p.Data = dataBytes
// 		reqBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		wsServer := &echoWebsocketServer{
// 			t:        t,
// 			response: [][]byte{[]byte(`{"type": "error", "topic": "availablesymbols", "event": "10", "data": {"message": "internal server error"}}`)},
// 			request:  reqBytes,
// 		}

// 		url := setupTestWebsocketServer(wsServer)

// 		client := dvotcWS.NewDVOTCClient(url, "123", "321")
// 		resp, err := client.PlaceLimitOrder(limitOrder)
// 		require.ErrorContains(t, err, "internal server error")

// 		require.Nil(t, resp)
// 	})
// }

// func TestCancellingOrder(t *testing.T) {
// 	t.Run("Successfully cancelled order", func(t *testing.T) {
// 		orderID := faker.UUIDHyphenated()
// 		p := dvotcWS.Payload{
// 			Type:  "request-response",
// 			Event: "10",
// 			Topic: fmt.Sprintf("cancelorder/%s", orderID),
// 		}
// 		reqBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		p.Data = []byte(`{"message": "Order successfully cancelled."}`)
// 		respBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		wsServer := &echoWebsocketServer{
// 			t:        t,
// 			response: [][]byte{respBytes},
// 			request:  reqBytes,
// 		}

// 		url := setupTestWebsocketServer(wsServer)

// 		client := dvotcWS.NewDVOTCClient(url, "123", "321")
// 		err = client.CancelOrder(orderID)
// 		require.NoError(t, err)
// 	})

// 	t.Run("Error cancelling order", func(t *testing.T) {
// 		orderID := faker.UUIDHyphenated()
// 		p := dvotcWS.Payload{
// 			Type:  "request-response",
// 			Event: "10",
// 			Topic: fmt.Sprintf("cancelorder/%s", orderID),
// 		}
// 		reqBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		p.Type = "error"
// 		p.Data = []byte(`{"message": "Failed to cancel order"}`)
// 		respBytes, err := json.Marshal(p)
// 		require.NoError(t, err)

// 		wsServer := &echoWebsocketServer{
// 			t:        t,
// 			response: [][]byte{respBytes},
// 			request:  reqBytes,
// 		}

// 		url := setupTestWebsocketServer(wsServer)

// 		client := dvotcWS.NewDVOTCClient(url, "123", "321")
// 		err = client.CancelOrder(orderID)
// 		require.ErrorContains(t, err, "Failed to cancel order")
// 	})
// }

// func TestSubscribeOrderChanges(t *testing.T) {
// 	p := dvotcWS.Payload{
// 		Type:  "subscribe",
// 		Topic: "order/#",
// 		Event: "order-updates",
// 	}

// 	var respData [2][]byte
// 	now := time.Now().UTC()
// 	data := dvotcWS.OrderStatus{}
// 	orderStatuses := []dvotcWS.OrderStatus{}
// 	err := faker.FakeData(&data, options.WithRandomMapAndSliceMaxSize(5))
// 	require.NoError(t, err)

// 	// order pending
// 	data.CancelledAt = nil
// 	data.FilledAt = nil
// 	data.CreatedAt = now
// 	data.Status = "Open"
// 	orderStatuses = append(orderStatuses, data)

// 	dataBytes, err := json.Marshal(data)
// 	require.NoError(t, err)
// 	p.Data = dataBytes
// 	respBytes, err := json.Marshal(p)
// 	require.NoError(t, err)
// 	respData[0] = respBytes

// 	// order complete
// 	now = now.Add(time.Second * 15)
// 	data.Status = "Complete"
// 	data.FilledAt = &now
// 	orderStatuses = append(orderStatuses, data)

// 	dataBytes, err = json.Marshal(data)
// 	require.NoError(t, err)
// 	p.Data = dataBytes
// 	respBytes, err = json.Marshal(p)
// 	require.NoError(t, err)
// 	respData[1] = respBytes

// 	wsServer := &echoV2WebsocketServer{
// 		t:      t,
// 		rrChan: make(chan [2][]byte),
// 	}

// 	url := setupTestV2WebsocketServer(wsServer)

// 	client := dvotcWS.NewDVOTCClient(url, "123", "321")
// 	sub, err := client.SubscribeOrderChanges("#")
// 	require.NoError(t, err)

// 	wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "order/#", "event": "order-updates"}`), respData[0]}
// 	d := <-sub.Data
// 	require.Equal(t, orderStatuses[0], d)

// 	wsServer.rrChan <- [2][]byte{nil, respData[1]}
// 	d = <-sub.Data
// 	require.Equal(t, orderStatuses[1], d)

// 	// shutting down server and client
// 	err = wsServer.StopServer()
// 	require.NoError(t, err)

// 	err = sub.StopConsuming()
// 	require.NoError(t, err)

// 	// produce error shutting down subscription twice
// 	err = sub.StopConsuming()
// 	require.ErrorIs(t, err, dvotcWS.ErrSubscriptionAlreadyClosed)
// }

// func TestSubscribeOrderChanges_Reconnect(t *testing.T) {
// 	p := dvotcWS.Payload{
// 		Type:  "subscribe",
// 		Topic: "order/#",
// 		Event: "order-updates",
// 	}

// 	var respData [2][]byte
// 	now := time.Now().UTC()
// 	data := dvotcWS.OrderStatus{}
// 	orderStatuses := []dvotcWS.OrderStatus{}
// 	err := faker.FakeData(&data, options.WithRandomMapAndSliceMaxSize(5))
// 	require.NoError(t, err)

// 	// order pending
// 	data.CancelledAt = nil
// 	data.FilledAt = nil
// 	data.CreatedAt = now
// 	data.Status = "Open"
// 	orderStatuses = append(orderStatuses, data)

// 	dataBytes, err := json.Marshal(data)
// 	require.NoError(t, err)
// 	p.Data = dataBytes
// 	respBytes, err := json.Marshal(p)
// 	require.NoError(t, err)
// 	respData[0] = respBytes

// 	// order complete
// 	now = now.Add(time.Second * 15)
// 	data.Status = "Complete"
// 	data.FilledAt = &now
// 	orderStatuses = append(orderStatuses, data)

// 	dataBytes, err = json.Marshal(data)
// 	require.NoError(t, err)
// 	p.Data = dataBytes
// 	respBytes, err = json.Marshal(p)
// 	require.NoError(t, err)
// 	respData[1] = respBytes

// 	wsServer := &echoV2WebsocketServer{
// 		t:      t,
// 		rrChan: make(chan [2][]byte),
// 	}

// 	url := setupTestV2WebsocketServer(wsServer)

// 	client := dvotcWS.NewDVOTCClient(url, "123", "321")
// 	sub, err := client.SubscribeOrderChanges("#")
// 	require.NoError(t, err)

// 	wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "order/#", "event": "order-updates"}`), respData[0]}
// 	d := <-sub.Data
// 	require.Equal(t, orderStatuses[0], d)

// 	wsServer.rrChan <- [2][]byte{nil, respData[1]}
// 	d = <-sub.Data
// 	require.Equal(t, orderStatuses[1], d)

// 	// reconnect message
// 	wsServer.rrChan <- [2][]byte{nil, []byte(`{"type": "info", "event": "reconnect"}`)}
// 	wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "order/#", "event": "order-updates"}`), respData[1]}
// 	d = <-sub.Data
// 	require.Equal(t, orderStatuses[1], d)

// 	// shutting down server and client
// 	err = wsServer.StopServer()
// 	require.NoError(t, err)

// 	err = sub.StopConsuming()
// 	require.NoError(t, err)

// 	// produce error shutting down subscription twice
// 	err = sub.StopConsuming()
// 	require.ErrorIs(t, err, dvotcWS.ErrSubscriptionAlreadyClosed)
// }
