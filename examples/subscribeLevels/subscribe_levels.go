package main

import (
	"fmt"
	"time"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
)

func main() {

	dvotcClient := dvotcWS.NewDVOTCClient("DVOTC_WS_URL", "YOUR_API_KEY", "YOUR_API_SECRET")
	err := dvotcClient.Ping()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	subscription, err := dvotcClient.SubscribeLevels("BTC/USD")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	go func() {
		for {
			data, ok := subscription.Data.Dequeue()
			if !ok {
				continue
			}
			quoteID := data.QuoteID
			levels := data.Levels
			fmt.Printf("quoteID: %s %+v\n", quoteID, levels)
			for _, l := range levels {
				fmt.Printf("%+v \n", l)
			}
		}
	}()

	time.Sleep(time.Second * 5)
}
