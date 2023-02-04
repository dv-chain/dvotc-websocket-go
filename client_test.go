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
	"strings"
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

func TestSettingUpConnectionAndValidatingSecret(t *testing.T) {
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

func TestSettingUpConnection_Fail(t *testing.T) {
	apiKey := faker.UUIDHyphenated()
	apiSecret := "das87d8sa7d98a7s89dhb"

	t.Run("invalid_ws_url", func(t *testing.T) {
		// w/o protocol
		client := dvotcWS.NewDVOTCClient("^WQE&^E^QW%E", apiKey, apiSecret)
		err := client.Ping()
		assert.ErrorContains(t, err, "invalid URL escape")
	})

	t.Run("fail_to_dial", func(t *testing.T) {
		client := dvotcWS.NewDVOTCClient("wss://something/websocket", apiKey, apiSecret)
		err := client.Ping()
		assert.ErrorContains(t, err, "dial tcp: lookup something: no such host")
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// setupTestWebsocketServer sets up websocket server and returns url to access
// request data passed in is asserted for actual value received
func setupTestWebsocketServer(e *echoWebsocketServer) string {
	srv := httptest.NewServer(http.HandlerFunc(e.handler))
	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"

	return u.String()
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

	defer func() {
		err := e.StopServer()
		require.NoError(e.t, err)
	}()
	mt, p, err := conn.ReadMessage()
	if err != nil {
		log.Printf("cannot read message: %v", err)
		return
	}
	fmt.Println("req", string(p))
	require.JSONEq(e.t, string(e.request), string(p))

	for _, m := range e.response {
		fmt.Println("resp", string(m))
		err := conn.WriteMessage(mt, m)
		require.NoError(e.t, err)
	}
}

func (e *echoWebsocketServer) StopServer() error {
	fmt.Println("end")
	return e.conn.Close()
}

func setupTestV2WebsocketServer(e *echoV2WebsocketServer) string {
	srv := httptest.NewServer(http.HandlerFunc(e.handler))
	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"

	return u.String()
}

type echoV2WebsocketServer struct {
	t          *testing.T
	conn       *websocket.Conn
	apiKey     string
	signature  string
	timestamp  string
	timeWindow string
	rrChan     chan ([2][]byte)
}

func (e *echoV2WebsocketServer) handler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("initiate connection ", len(e.rrChan))
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot upgrade: %v", err), http.StatusInternalServerError)
	}

	e.apiKey = req.Header.Get("dv-api-key")
	e.signature = req.Header.Get("dv-signature")
	e.timestamp = req.Header.Get("dv-timestamp")
	e.timeWindow = req.Header.Get("dv-timewindow")
	e.conn = conn

	count := 1
	for rr := range e.rrChan {
		req, res := rr[0], rr[1]
		// var mt int
		// var p []byte
		// var err error
		fmt.Println("writing count ", count)
		fmt.Println(string(req), string(res))
		if len(req) > 0 {
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Printf("cannot read message: %v", err)
				return
			}
			fmt.Println("validating ", string(req), string(p))
			require.JSONEq(e.t, string(req), string(p))
		}

		if len(res) > 0 {
			if err := conn.WriteMessage(1, res); err != nil {
				log.Printf("cannot wrtite message: %v", err)
				return
			}

			if strings.Contains(string(res), "reconnect") {
				fmt.Println("server shutting down due to reconnect")
				return
			}
		}
		count++
	}
}

func (e *echoV2WebsocketServer) StopServer() error {
	fmt.Println("end")
	return e.conn.Close()
}
