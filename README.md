# Long Bridge Broker Golang API

Long Bridge is a promising broker emerging in the market with first rate trading app. Recently (May 10, 2022) they have released the OpenAPI for quants https://longbridgeapp.com/en/topics/2543341?channel=t2543341&invite-code=0VMCD7.

This is a go implementation of full long bridge broker API based on the document in https://open.longbridgeapp.com/en/docs by the quantitive trading team from Deep Learning Limited.

For API in other languages, please refer to longbridge's official github site at https://github.com/longbridgeapp.

## Features

- Support full long bridge APIs to access trading, quote, portfolio and real time subscription.
- Maintain auto reconnection of long connection for trading, quote and subscriptions.
- Support connection through both secured web socket and TCP.
- Strong typed APIs.
- Pure golang implementation.


## Installation
    go get github.com/deepln-io/longbridge-goapi

## How to use it

The package provides two sets of APIs through **TradeClient** and **QuoteClient** for trading and quote respectively.
The library automatically manages the connection to long bridge servers and will reconnect if disconnected. This is useful for real production systems when a trading bot is running continuously overnight.

To use the trade notification, set order notification call back through **TradeClient.OnOrderChange**(order *longbridge.Order).

Note to run the examples, please make sure that your long bridge account is ready with access token, app key and secret available.
Set them in env variable before running.


## Example

Get the stock static information.

```
package main
import (
    "log"
    "os"

    "github.com/deepln-io/longbridge-goapi"
)

func main() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",  // Optionally choose TCP connection here
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ss, err := c.GetStaticInfo([]string{"700.HK", "AAPL.US"})
	if err != nil {
		log.Fatalf("Error getting static info: %v", err)
	}
	for _, s := range ss {
		log.Printf("%#v\n", s)
	}
}
```

Get account positions
```go
package main
import (
    "log"
    "os"

    "github.com/deepln-io/longbridge-goapi"
)

func main() {
c, err := longbridge.NewTradeClient(&longbridge.Config{
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	positions, err := c.GetStockPositions()
	if err != nil {
		log.Fatalf("Error getting stock positions: %v", err)
	}
	for _, position := range positions {
		log.Printf("%+v", position)
	}
}
```

Place an order

```go
package main
import (
    "log"
    "os"

    "github.com/deepln-io/longbridge-goapi"
)

func main() {
        c, err := longbridge.NewTradeClient(&longbridge.Config{
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	orderID, err := c.PlaceOrder(&longbridge.PlaceOrderReq{
		Symbol:      "AMD.US",
		OrderType:   longbridge.LimitOrder,
		Price:       180,
		Quantity:    1,
		ExpireDate:  time.Now().AddDate(0, 1, 0),
		Side:        longbridge.Sell,
		OutsideRTH:  longbridge.AnyTime,
		TimeInForce: longbridge.GoodTilCancel,
		Remark:      "钓鱼单",
	})
	if err != nil {
		log.Fatalf("Error placing order: %v", err)
	}
```


## Acknowledge

Thanks the long bridge development team for patiently replying a lot of technical questions to clarify the API details during our development.

## License

MIT License

## Contributing

We are always happy to welcome new contributors! If you have any questions, please feel free to reach out by opening an issue or leaving a comment.