package dvotcWS_test

import (
	"encoding/json"
	"testing"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListLimitsBalances(t *testing.T) {
	p := dvotcWS.Payload{
		Type:  "request-response",
		Topic: "limits",
		Event: "10",
	}
	data := dvotcWS.AssetBalance{
		Assets:     []dvotcWS.Asset{},
		UsdBalance: 23123,
	}

	i := 0
	for i < 5 {
		asset := dvotcWS.Asset{}
		err := faker.FakeData(&asset)
		require.NoError(t, err)
		data.Assets = append(data.Assets, asset)
		i += 1
	}

	dataBytes, err := json.Marshal(data)
	require.NoError(t, err)
	p.Data = dataBytes

	respBytes, err := json.Marshal(p)
	require.NoError(t, err)

	url := setupTestWebsocketServer(
		&echoWebsocketServer{
			t:        t,
			request:  []byte(`{"type": "request-response", "topic": "limits", "event": "10"}`),
			response: [][]byte{respBytes},
		},
	)

	client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
	assets, err := client.ListLimitsBalances()
	assert.NoError(t, err)
	assert.Len(t, assets.Assets, 5)
}

func TestListLimitsBalances_Error(t *testing.T) {
	url := setupTestWebsocketServer(
		&echoWebsocketServer{
			t:        t,
			request:  []byte(`{"type": "request-response", "topic": "availablesymbols", "event": "10"}`),
			response: [][]byte{[]byte(`{"type": "error", "topic": "availablesymbols", "event": "10", "data": {"message": "internal server error"}}`)},
		},
	)

	client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
	symbols, err := client.ListAvailableSymbols()
	assert.EqualError(t, err, `{"message": "internal server error"}`)
	assert.Len(t, symbols, 0)
	assert.NotEqualValues(t, []string{"XRP/USD", "XRP/CAD"}, symbols)
}
