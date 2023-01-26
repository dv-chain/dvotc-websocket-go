DVOTC Websocket API in Go
===============
This package helps execute trade/orders against DVChains websocket API in golang!

## Installation

### *go get*
```sh
$ go get -u github.com/dv-chain/dvotc-websocket-api
```


## Example

More examples can be  found in `/examples` folder

```golang
package main

import (
	"fmt"

	dvotcWS "github.com/dv-chain/dvotc-websocket-go"
)

func main() {
	// DVOTC_WS_URL = "sandbox.trade.dvchain.co" || "trade.dvchain.co"
	// YOUR_API_KEY = "4f8f48ff-3135-422c-9ce7-1cc5a31a72d8"
	// YOUR_API_SECRET = "n43n2423423nm4b4b34n32423"

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
```
