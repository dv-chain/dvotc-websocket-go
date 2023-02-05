package dvotcWS_test

import (
	"encoding/json"
	"fmt"
	"testing"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/* LOGIN */
func TestLoginNotification(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := dvotcWS.Payload{
			Type:  "subscribe",
			Event: "notifications",
			Topic: "LOGIN",
		}

		loginNotif := dvotcWS.LoginNotification{}
		err := faker.FakeData(&loginNotif, options.WithFieldsToIgnore("GroupAccount"))
		require.NoError(t, err)
		fmt.Println(loginNotif)

		loginNotif.User.GroupAccount = nil
		dataBytes, err := json.Marshal(loginNotif)
		require.NoError(t, err)
		p.Data = dataBytes

		respBytes, err := json.Marshal(p)
		require.NoError(t, err)

		wsServer := &echoV2WebsocketServer{
			t:      t,
			rrChan: make(chan [2][]byte),
		}

		url := setupTestV2WebsocketServer(wsServer)

		client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
		sub, err := client.SubscribeLogin()
		assert.NoError(t, err)

		wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "LOGIN", "event": "notifications"}`), respBytes}
		notif := <-sub.Data

		assert.Equal(t, loginNotif, notif)

		assert.NoError(t, wsServer.StopServer())
		assert.NoError(t, sub.StopConsuming())
	})

	t.Run("reconnect_success", func(t *testing.T) {
		p := dvotcWS.Payload{
			Type:  "subscribe",
			Event: "notifications",
			Topic: "LOGIN",
		}

		loginNotif := dvotcWS.LoginNotification{}
		err := faker.FakeData(&loginNotif, options.WithFieldsToIgnore("GroupAccount"))
		require.NoError(t, err)
		fmt.Println(loginNotif)

		loginNotif.User.GroupAccount = nil
		dataBytes, err := json.Marshal(loginNotif)
		require.NoError(t, err)
		p.Data = dataBytes

		respBytes, err := json.Marshal(p)
		require.NoError(t, err)

		wsServer := &echoV2WebsocketServer{
			t:      t,
			rrChan: make(chan [2][]byte),
		}

		url := setupTestV2WebsocketServer(wsServer)

		client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
		sub, err := client.SubscribeLogin()
		assert.NoError(t, err)

		wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "LOGIN", "event": "notifications"}`), respBytes}
		notif := <-sub.Data
		assert.Equal(t, loginNotif, notif)

		// reconnect
		wsServer.rrChan <- [2][]byte{nil, []byte(`{"type": "info", "event": "reconnect"}`)}
		// resend initial message
		wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "LOGIN", "event": "notifications"}`), respBytes}
		notif = <-sub.Data
		assert.Equal(t, loginNotif, notif)

		assert.NoError(t, wsServer.StopServer())
		assert.NoError(t, sub.StopConsuming())
	})

	t.Run("reconnect_fail", func(t *testing.T) {
		p := dvotcWS.Payload{
			Type:  "subscribe",
			Event: "notifications",
			Topic: "LOGIN",
		}

		loginNotif := dvotcWS.LoginNotification{}
		err := faker.FakeData(&loginNotif, options.WithFieldsToIgnore("GroupAccount"))
		require.NoError(t, err)
		fmt.Println(loginNotif)

		loginNotif.User.GroupAccount = nil
		dataBytes, err := json.Marshal(loginNotif)
		require.NoError(t, err)
		p.Data = dataBytes

		respBytes, err := json.Marshal(p)
		require.NoError(t, err)

		wsServer := &echoV2WebsocketServer{
			t:      t,
			rrChan: make(chan [2][]byte),
		}

		url := setupTestV2WebsocketServer(wsServer)

		client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
		sub, err := client.SubscribeLogin()
		assert.NoError(t, err)

		wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "LOGIN", "event": "notifications"}`), respBytes}
		notif := <-sub.Data
		assert.Equal(t, loginNotif, notif)

		// reconnect message and shut down server
		wsServer.rrChan <- [2][]byte{nil, []byte(`{"type": "info", "event": "reconnect"}`)}
		assert.NoError(t, wsServer.StopServer())

		// channel closed to empty message sent
		notif = <-sub.Data
		assert.Equal(t, dvotcWS.LoginNotification{}, notif)
		assert.NoError(t, sub.StopConsuming())
	})

	t.Run("server_closed_unexpectedly", func(t *testing.T) {
		p := dvotcWS.Payload{
			Type:  "subscribe",
			Event: "notifications",
			Topic: "LOGIN",
		}

		loginNotif := dvotcWS.LoginNotification{}
		err := faker.FakeData(&loginNotif, options.WithFieldsToIgnore("GroupAccount"))
		require.NoError(t, err)
		fmt.Println(loginNotif)

		loginNotif.User.GroupAccount = nil
		dataBytes, err := json.Marshal(loginNotif)
		require.NoError(t, err)
		p.Data = dataBytes

		respBytes, err := json.Marshal(p)
		require.NoError(t, err)

		wsServer := &echoV2WebsocketServer{
			t:      t,
			rrChan: make(chan [2][]byte),
		}

		url := setupTestV2WebsocketServer(wsServer)

		client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
		sub, err := client.SubscribeLogin()
		assert.NoError(t, err)

		wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "LOGIN", "event": "notifications"}`), respBytes}
		notif := <-sub.Data
		assert.Equal(t, loginNotif, notif)

		assert.NoError(t, wsServer.StopServer())

		// channel closed
		notif = <-sub.Data
		assert.Equal(t, dvotcWS.LoginNotification{}, notif)
		assert.NoError(t, sub.StopConsuming())
	})
}

/* BATCH_CREATED */
func TestBatchCreatedNotification(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := dvotcWS.Payload{
			Type:  "subscribe",
			Event: "notifications",
			Topic: "BATCH_CREATED",
		}

		batchCreatedNotif := dvotcWS.BatchCreatedNotification{}
		err := faker.FakeData(&batchCreatedNotif, options.WithFieldsToIgnore("GroupAccount"), options.WithRandomMapAndSliceMaxSize(2), options.WithRandomMapAndSliceMinSize(2))
		require.NoError(t, err)
		fmt.Println(batchCreatedNotif)

		batchCreatedNotif.User.GroupAccount = nil
		dataBytes, err := json.Marshal(batchCreatedNotif)
		require.NoError(t, err)
		p.Data = dataBytes

		respBytes, err := json.Marshal(p)
		require.NoError(t, err)

		wsServer := &echoV2WebsocketServer{
			t:      t,
			rrChan: make(chan [2][]byte),
		}

		url := setupTestV2WebsocketServer(wsServer)

		client := dvotcWS.NewDVOTCClient(url+"/websocket", "123", "321")
		sub, err := client.SubscribeBatchCreated()
		assert.NoError(t, err)

		wsServer.rrChan <- [2][]byte{[]byte(`{"type": "subscribe", "topic": "BATCH_CREATED", "event": "notifications"}`), respBytes}
		notif := <-sub.Data
		assert.Equal(t, batchCreatedNotif, notif)

		assert.NoError(t, wsServer.StopServer())
		assert.NoError(t, sub.StopConsuming())
	})
}
