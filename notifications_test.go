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

func TestLoginNotification(t *testing.T) {
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
}
