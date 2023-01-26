package main

import (
	"fmt"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
)

func main() {
	dvotcClient := dvotcWS.NewDVOTCClient("DVOTC_WS_URL", "YOUR_API_KEY", "YOUR_API_SECRET")
	err := dvotcClient.Ping()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	symbols, err := dvotcClient.ListAvailableSymbols()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	for i, symbols := range symbols {
		fmt.Printf("%d) %s", i, symbols)
	}
}
