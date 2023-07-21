package dvotcWS_test

import (
	"testing"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
	"github.com/stretchr/testify/assert"
)

func TestListAvailableSymbols(t *testing.T) {

	url := setupTestWebsocketServer(
		&echoWebsocketServer{
			t:        t,
			request:  []byte(`{"type": "request-response", "topic": "availablesymbols", "event": "10"}`),
			response: [][]byte{[]byte(`{"type": "request-response", "topic": "availablesymbols", "event": "10", "data": ["BTC/USD", "BTC/CAD"]}`)},
		},
	)

	client := dvotcWS.NewDVOTCClient(url, "123", "321")
	symbols, err := client.ListAvailableSymbols()
	assert.NoError(t, err)
	assert.Len(t, symbols, 2)
	assert.EqualValues(t, []string{"BTC/USD", "BTC/CAD"}, symbols)
	assert.NotEqualValues(t, []string{"XRP/USD", "XRP/CAD"}, symbols)
}

func TestListAvailableSymbols_Error(t *testing.T) {
	url := setupTestWebsocketServer(
		&echoWebsocketServer{
			t:        t,
			request:  []byte(`{"type": "request-response", "topic": "availablesymbols", "event": "10"}`),
			response: [][]byte{[]byte(`{"type": "error", "topic": "availablesymbols", "event": "10", "data": {"message": "internal server error"}}`)},
		},
	)

	client := dvotcWS.NewDVOTCClient(url, "123", "321")
	symbols, err := client.ListAvailableSymbols()
	assert.EqualError(t, err, `{"message": "internal server error"}`)
	assert.Len(t, symbols, 0)
	assert.NotEqualValues(t, []string{"XRP/USD", "XRP/CAD"}, symbols)
}
