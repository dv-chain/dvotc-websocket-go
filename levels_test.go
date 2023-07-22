package dvotcWS_test

import (
	"encoding/json"
	"testing"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/stretchr/testify/require"
)

func TestListLevels(t *testing.T) {
	p := dvotcWS.Payload{
		Type:  "subscribe",
		Topic: "BTC/USD",
		Event: "levels",
	}

	var respData [5][]byte
	var i int = 0
	var levelData []*dvotcWS.LevelData
	for i < 5 {
		data := &dvotcWS.LevelData{}
		err := faker.FakeData(&data, options.WithRandomMapAndSliceMaxSize(5))
		require.NoError(t, err)
		levelData = append(levelData, data)

		dataBytes, err := json.Marshal(data)
		require.NoError(t, err)
		p.Data = dataBytes
		respBytes, err := json.Marshal(p)
		require.NoError(t, err)
		respData[i] = respBytes
		i++
	}

	wsServer := &echoV2WebsocketServer{
		t:      t,
		rrChan: make(chan [2][]byte),
	}

	url := setupTestV2WebsocketServer(wsServer)

	client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
	sub, err := client.SubscribeLevels("BTC/USD")
	require.NoError(t, err)

	wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "BTC/USD", "event": "levels"}`), respData[0]}
	d := <-sub.Data
	require.Equal(t, levelData[0], d)

	wsServer.rrChan <- [2][]byte{nil, respData[1]}
	d = <-sub.Data
	require.Equal(t, levelData[1], d)

	// reconnect message
	wsServer.rrChan <- [2][]byte{nil, []byte(`{"type": "info", "event": "reconnect"}`)}
	wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "BTC/USD", "event": "levels"}`), respData[2]}
	d = <-sub.Data
	require.Equal(t, levelData[2], d)

	// shutting down server and client
	err = wsServer.StopServer()
	require.NoError(t, err)

	err = sub.StopConsuming()
	require.NoError(t, err)

	// produce error shutting down subscription twice
	err = sub.StopConsuming()
	require.ErrorIs(t, err, dvotcWS.ErrSubscriptionAlreadyClosed)
}

func TestListLevels_ReuseConnection(t *testing.T) {
	p := dvotcWS.Payload{
		Type:  "subscribe",
		Topic: "BTC/USD",
		Event: "levels",
	}

	var respData [5][]byte
	var i int = 0
	var levelData []*dvotcWS.LevelData
	for i < 3 {
		data := &dvotcWS.LevelData{}
		err := faker.FakeData(&data, options.WithRandomMapAndSliceMaxSize(5))
		require.NoError(t, err)
		levelData = append(levelData, data)

		dataBytes, err := json.Marshal(data)
		require.NoError(t, err)
		p.Data = dataBytes
		respBytes, err := json.Marshal(p)
		require.NoError(t, err)
		respData[i] = respBytes
		i++
	}

	wsServer := &echoV2WebsocketServer{
		t:      t,
		rrChan: make(chan [2][]byte),
	}

	url := setupTestV2WebsocketServer(wsServer)

	client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
	sub, err := client.SubscribeLevels("BTC/USD")
	require.NoError(t, err)

	sub2, err := client.SubscribeLevels("BTC/USD")
	require.NoError(t, err)

	wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "BTC/USD", "event": "levels"}`), respData[0]}
	d := <-sub.Data
	require.Equal(t, levelData[0], d)
	d2 := <-sub2.Data
	require.Equal(t, levelData[0], d2)

	wsServer.rrChan <- [2][]byte{nil, respData[1]}
	d = <-sub.Data
	require.Equal(t, levelData[1], d)
	d2 = <-sub2.Data
	require.Equal(t, levelData[1], d2)

	// reconnect message
	wsServer.rrChan <- [2][]byte{nil, []byte(`{"type": "info", "event": "reconnect"}`)}
	wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "BTC/USD", "event": "levels"}`), respData[2]}
	d = <-sub.Data
	require.Equal(t, levelData[2], d)
	d2 = <-sub2.Data
	require.Equal(t, levelData[2], d2)

	// shutting down server and client
	err = wsServer.StopServer()
	require.NoError(t, err)

	err = sub.StopConsuming()
	require.NoError(t, err)

	// produce error shutting down subscription twice
	err = sub2.StopConsuming()
	require.NoError(t, err)
}
