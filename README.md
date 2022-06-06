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

The API is designed like a file system. Call **longbridge.New()** to create a file handler-like client. Call **client.Close()** after use. For using HTTP APIs to place new orders or view history orders, just call the exported methods **client.PlaceOrder**, **GetHistoryOrders**, etc.

To use the quote services, set **client.Quote.Enable(true)** and call the exported functions on Quote to access the services. The library automatically manages the connection to long bridge servers and will reconnect if disconnected. This is useful for real production systems when a trading bot is running continuously overnight.

To use the trade notification, set **client.Trade.Enable(true)** and set order notification call back through **client.Trade.OnOrderChange**(order *longbridge.Order).

Note to run the examples, please make sure that your long bridge account is ready with access token, app key and secret available.
Set them in env variable before running.


## Example

Place an order

```go
package main
import (
    "log"

    "github.com/deepln-io/longbridge-goapi"
)

func main() {
	c, err := longbridge.NewClient(&Config{
		BaseURL:     "https://openapi.longbridgeapp.com",
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	if err := c.PlaceOrder(&longbridge.PlaceOrderReq{
    	Symbol: "AAPL.US",
    	OrderType: longbridge.EnhancedLimitOrder,
    	Price: 130.0,
    	Quantity:1,
    	Side: longbridge.Buy,
    	TimeInForce: longbridge.DayOrder,
    	Remark: "An apple a day keeps the doctor away",
	}); err !=nil {
	    log.Fatalf("Error placing buy order for Apple: %v", err)
	}
}
```

Get the stock static information.

```go
package main
import (
    "log"

    "github.com/deepln-io/longbridge-goapi"
)

func main() {
	c, err := longbridge.NewClient(&Config{
		BaseURL:     "https://openapi.longbridgeapp.com",
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	c.Quote.Enable(true)
	ss, err := c.Quote.GetStaticInfo([]string{"700.HK", "AAPL.US"})
	if err != nil {
		log.Fatalf("Error getting static info: %v", err)
	}
	for _, s := range ss {
		log.Printf("%#v\n", s)
	}
}
```

## Acknowledge

Thanks the long bridge development team for patiently replying a lot of technical questions to clarify the API details during our development.

## Contributing

We are always happy to welcome new contributors! If you have any questions, please feel free to reach out by opening an issue or leaving a comment.