package dvotcWS_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
	"github.com/fasthttp/websocket"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Verify(msg, key []byte, hash string) (bool, error) {
	sig, err := base64.StdEncoding.DecodeString(hash)
	// sig, err := hex.DecodeString(hash)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(msg)

	return hmac.Equal(sig, mac.Sum(nil)), nil
}

func TestSettingUpConnection(t *testing.T) {
	e := &echoWebsocketServer{
		t:        t,
		request:  []byte(`{"type": "ping-pong", "topic": "Topiccs", "event": "10"}`),
		response: [][]byte{[]byte(`{"type": "ping-pong", "topic": "Topiccs", "event": "10"}`)},
	}
	url := setupTestWebsocketServer(e)

	apiKey := faker.UUIDHyphenated()
	apiSecret := "das87d8sa7d98a7s89dhb"

	client := dvotcWS.NewDVOTCClient(url+"/websocket", apiKey, apiSecret)
	err := client.Ping()
	assert.NoError(t, err)

	// verify signature
	assert.Equal(t, apiKey, e.apiKey)
	msg := fmt.Sprintf("%s%s%s", e.apiKey, e.timestamp, e.timeWindow)
	isValid, err := Verify([]byte(msg), []byte(apiSecret), e.signature)
	assert.NoError(t, err)
	assert.True(t, isValid, "websocket signature not valid")
}

func setupTestWebsocketServer(e *echoWebsocketServer) string {
	srv := httptest.NewServer(http.HandlerFunc(e.handler))
	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"

	return u.String()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type echoWebsocketServer struct {
	t          *testing.T
	conn       *websocket.Conn
	apiKey     string
	signature  string
	timestamp  string
	timeWindow string
	response   [][]byte
	request    []byte
}

func (e *echoWebsocketServer) handler(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot upgrade: %v", err), http.StatusInternalServerError)
	}
	e.conn = conn

	e.apiKey = req.Header.Get("dv-api-key")
	e.signature = req.Header.Get("dv-signature")
	e.timestamp = req.Header.Get("dv-timestamp")
	e.timeWindow = req.Header.Get("dv-timewindow")

	defer e.StopServer()
	mt, p, err := conn.ReadMessage()
	if err != nil {
		log.Printf("cannot read message: %v", err)
		return
	}
	require.JSONEq(e.t, string(e.request), string(p))

	for _, m := range e.response {
		conn.WriteMessage(mt, m)
	}
}

func (e *echoWebsocketServer) StopServer() error {
	return e.conn.Close()
}
