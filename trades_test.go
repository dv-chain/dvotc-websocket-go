package dvotcWS_test

import (
	"encoding/json"
	"testing"
	"time"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTrades(t *testing.T) {
	tradeParms := dvotcWS.ListTradesPayload{}
	b, err := json.Marshal(tradeParms)
	require.NoError(t, err)
	p := dvotcWS.Payload{
		Type:  "request-response",
		Topic: "tradestatus",
		Event: "10",
		Data:  b,
	}
	reqBytes, err := json.Marshal(p)
	require.NoError(t, err)

	tradesResp := []dvotcWS.Trade{}
	i := 0
	for i < 5 {
		trade := dvotcWS.Trade{}
		err := faker.FakeData(&trade)
		assert.NoError(t, err)
		t := time.Now().UTC()
		trade.FilledAt = t
		trade.CreatedAt = t
		trade.UpdatedAt = &t
		tradesResp = append(tradesResp, trade)
		i++
	}

	dataBytes, err := json.Marshal(tradesResp)
	require.NoError(t, err)
	p.Data = dataBytes
	respBytes, err := json.Marshal(p)
	require.NoError(t, err)

	url := setupTestWebsocketServer(
		&echoWebsocketServer{
			t:        t,
			request:  reqBytes,
			response: [][]byte{[]byte(respBytes)},
		},
	)

	client := dvotcWS.NewDVOTCClient(url, "123", "321")
	trades, err := client.ListTrades(nil, nil, nil)
	assert.NoError(t, err)
	assert.Len(t, trades, i)
	assert.Equal(t, tradesResp, trades)
}

func TestListTrades_Error(t *testing.T) {
	tradeParms := dvotcWS.ListTradesPayload{}
	b, err := json.Marshal(tradeParms)
	require.NoError(t, err)
	p := dvotcWS.Payload{
		Type:  "request-response",
		Topic: "tradestatus",
		Event: "10",
		Data:  b,
	}
	reqBytes, err := json.Marshal(p)
	require.NoError(t, err)

	respPayload := dvotcWS.Payload{
		Type:  "error",
		Topic: "tradestatus",
		Event: "10",
		Data:  []byte(`{"message":"internal server error"}`),
	}
	respBytes, err := json.Marshal(respPayload)
	require.NoError(t, err)

	url := setupTestWebsocketServer(
		&echoWebsocketServer{
			t:        t,
			request:  reqBytes,
			response: [][]byte{[]byte(respBytes)},
		},
	)

	client := dvotcWS.NewDVOTCClient(url, "123", "321")
	trades, err := client.ListTrades(nil, nil, nil)
	assert.EqualError(t, err, `{"message":"internal server error"}`)
	assert.Len(t, trades, 0)
}
