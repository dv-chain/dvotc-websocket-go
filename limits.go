package dvotcWS

import (
	"encoding/json"
	"fmt"
)

type AssetBalance struct {
	Assets     []Asset `json:"assets"`
	UsdBalance float64 `json:"usdBalance"`
}

type Asset struct {
	Asset    string  `json:"asset"`
	MaxSell  float64 `json:"maxSell"`
	MaxBuy   float64 `json:"maxBuy"`
	Position float64 `json:"position"`
}

func (dvotc *DVOTCClient) ListLimitsBalances() (*AssetBalance, error) {
	conn, err := dvotc.getConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	payload := Payload{
		Type:  MessageTypeRequestResponse,
		Event: dvotc.getRequestID(),
		Topic: "limits",
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	resp := &Payload{}
	if err := conn.ReadJSON(resp); err != nil {
		return nil, err
	}
	if resp.Type == "error" {
		return nil, fmt.Errorf("%s", string(resp.Data))
	}

	assetBalances := AssetBalance{}
	if err := json.Unmarshal(resp.Data, &assetBalances); err != nil {
		return nil, err
	}

	return &assetBalances, nil
}
