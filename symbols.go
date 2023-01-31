package dvotcWS

import (
	"encoding/json"
	"fmt"
)

func (dvotc *DVOTCClient) ListAvailableSymbols() ([]string, error) {
	conn, err := dvotc.getConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	payload := Payload{
		Type:  MessageTypeRequestResponse,
		Event: dvotc.getRequestID(),
		Topic: "availablesymbols",
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	resp := Payload{}
	err = conn.ReadJSON(&resp)
	if err != nil {
		return nil, err
	}
	if resp.Type == "error" {
		return nil, fmt.Errorf(string(resp.Data))
	}

	var symbols []string
	err = json.Unmarshal(resp.Data, &symbols)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}
