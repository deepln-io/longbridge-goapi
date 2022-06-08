# Long Bridge Broker Golang API

Long Bridge is a promising broker emerging in the stock market with first rate trading app. Recently (May 10, 2022) they released the OpenAPI for quants https://longbridgeapp.com/en/topics/2543341?channel=t2543341&invite-code=0VMCD7.

This is a go implementation of full long bridge broker APIs based on the document in https://open.longbridgeapp.com/en/docs.

For API in other languages, please refer to longbridge's official github site at https://github.com/longbridgeapp.

## Features

- Support full long bridge APIs for trading, accessing to account portfolio and subscribing real time market data.
- Strong typed APIs.
- Maintain auto reconnection of long connection for trading and quote subscriptions to fit for full automatic production environment.
- Support both secured web socket and TCP connection.
- Pure golang implementation.


## Installation
    go get github.com/deepln-io/longbridge-goapi

## How to use it

The package provides two sets of APIs through **TradeClient** and **QuoteClient** for trading and quote respectively.

The library automatically manages the connection to long bridge servers and will reconnect if disconnected. This is useful for real production systems when a trading bot is running continuously overnight.

To use the trade notification, set order notification call back through **TradeClient.OnOrderChange**(order *longbridge.Order).

*Note to run the examples, please make sure that your long bridge account is ready with access token, app key and secret available.*
Set them in env variable before running.
```sh
export LB_ACCESS_TOKEN="<access token from your account's developer center>"
export LB_APP_KEY="<your app key>"
export LB_APP_SECRET="<your app secret>"
```

## Examples

**Get the stock static information.**
The code connects to long bridge, gets down the static security information, and prints them out.
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

**Get real time price**
Make sure you have real time quote access permission from Long Bridge.
```go
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
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	qs, err := c.GetRealtimeQuote([]string{"5.HK", "MSFT.US"})
	if err != nil {
		log.Fatalf("Error getting real time quote: %v", err)
	}
	for _, q := range qs {
		log.Printf("%#v\n", q)
	}
}
```

**Get account positions**
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

**Place an order**

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

**Place, modify and cancel order**
```go
package main
import (
    "log"
    "os"
	"time"

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
	time.Sleep(time.Second)
	if err := c.ModifyOrder(&longbridge.ModifyOrderReq{
		OrderID:      orderID,
		Quantity:     1,
		Price:        200,
		TriggerPrice: 200,
	}); err != nil {
		log.Printf("Error modifying submitted order (id: %v): %v", orderID, err)
	}
	log.Printf("Order modified successfully, ID: %v", orderID)

	time.Sleep(time.Second)
	if err := c.CancelOrder(orderID); err != nil {
		log.Printf("Error cancelling submitted order (id: %v): %v", orderID, err)
	}

	log.Printf("Order cancelled successfully, ID: %v", orderID)
}
```

**Pull real time tick data**
```go
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
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ts, err := c.GetTickers("AAPL.US", 10)
	if err != nil {
		log.Fatalf("Error getting tickers for: %v", err)
	}
	for _, t := range ts {
		log.Printf("%#v", t)
	}
}
```

**Subscribe level 2 tick data with callback**
```go
package main
import (
    "log"
    "os"
	"time"

    "github.com/deepln-io/longbridge-goapi"
)

func main() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	c.OnPushTickers = func(t *longbridge.PushTickers) {
		log.Printf("Got tickers for %s, seq=%d", t.Symbol, t.Sequence)
		for _, ticker := range t.Tickers {
			log.Printf("Ticker: %#v", ticker)
		}
	}

	ss, err := c.SubscribePush([]string{"AAPL.US"},
		[]longbridge.SubscriptionType{longbridge.SubscriptionTicker}, true)
	if err != nil {
		log.Fatalf("Error subscribe quote push: %v", err)
	}
	for _, s := range ss {
		log.Printf("%#v", s)
	}

	time.Sleep(30 * time.Second)
	if err := c.UnsubscribePush(nil,
		[]longbridge.SubscriptionType{longbridge.SubscriptionTicker}, true); err != nil {
		log.Fatalf("Error unsubscribing all symbols' quote push: %v", err)
	}
}

```

## Acknowledge

Thanks the long bridge development team for patiently replying a lot of technical questions to clarify the API details during our development.

## License

[MIT License](https://opensource.org/licenses/MIT)

## Development

When longbridge protobuf files are updated, run the script gen-pb.sh in internal dir and push the changes.

## Contributing

We are always happy to welcome new contributors! If you have any questions, please feel free to reach out by opening an issue or leaving a comment.