// Copyright 2022 Deep Learning Limited. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package longbridge_test

import (
	"log"
	"os"
	"time"

	"github.com/deepln-io/longbridge-goapi"
)

func ExampleTradeClient_GetStockPositions() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
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

	// Output:
}

func ExampleTradeClient_GetFundPositions() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	positions, err := c.GetFundPositions()
	if err != nil {
		log.Fatalf("Error getting fund positions: %v", err)
	}
	for _, position := range positions {
		log.Printf("%+v", position)
	}

	// Output:
}

func ExampleTradeClient_GetAccountBalances() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	balances, err := c.GetAccountBalances()
	if err != nil {
		log.Fatalf("Error getting account balance: %v", err)
	}
	for _, b := range balances {
		log.Printf("%+v", b)
	}

	// Output:
}

func ExampleTradeClient_GetCashflows() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	cs, err := c.GetCashflows(&longbridge.CashflowReq{
		StartTimestamp: time.Now().AddDate(-1, 0, 0).Unix(),
		EndTimestamp:   time.Now().Unix()})
	if err != nil {
		log.Fatalf("Error getting account cash flow: %v", err)
	}
	for _, c := range cs {
		log.Printf("%+v\n", c)
	}

	// Output:
}

func ExampleTradeClient_PlaceOrder() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
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
	log.Printf("Order submitted successfully, ID: %v", orderID)
	// Output:
}

func ExampleTradeClient_CancelOrder() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
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
	if err := c.CancelOrder(orderID); err != nil {
		log.Fatalf("Error cancelling submitted order (id: %v): %v", orderID, err)
	}

	log.Printf("Order cancelled successfully, ID: %v", orderID)
	// Output:
}

func ExampleTradeClient_ModifyOrder() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
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
		log.Fatalf("Error modifying submitted order (id: %v): %v", orderID, err)
	}

	log.Printf("Order modified successfully, ID: %v", orderID)
	// Output:
}

func ExampleTradeClient_GetHistoryOrders() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	os, err := c.GetHistoryOrders(&longbridge.GetHistoryOrdersReq{
		Symbol:         "",
		Status:         nil,
		Side:           "",
		Market:         longbridge.HK,
		StartTimestamp: time.Now().AddDate(-1, 0, 0).Unix(),
		EndTimestamp:   time.Now().Unix(),
	})
	if err != nil {
		log.Fatalf("Error getting history orders: %v", err)
	}
	for _, o := range os {
		log.Printf("%+v\n", o)
	}

	// Output:
}

func ExampleTradeClient_GetTodayOrders() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	os, err := c.GetTodayOrders(&longbridge.GetTodyOrdersReq{
		Symbol:  "",
		Status:  nil,
		Side:    "",
		Market:  longbridge.US,
		OrderID: 0,
	})
	if err != nil {
		log.Fatalf("Error getting today's orders: %v", err)
	}
	for _, o := range os {
		log.Printf("%+v\n", o)
	}

	// Output:
}

func ExampleTradeClient_GetHistoryOrderFills() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	os, err := c.GetHistoryOrderFills(&longbridge.GetHistoryOrderFillsReq{
		Symbol:         "",
		StartTimestamp: time.Now().AddDate(-1, 0, 0).Unix(),
		EndTimestamp:   time.Now().Unix(),
	})
	if err != nil {
		log.Fatalf("Error getting history order fills: %v", err)
	}
	for _, o := range os {
		log.Printf("%+v\n", o)
	}

	// Output:
}

func ExampleTradeClient_GetTodayOrderFills() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	os, err := c.GetTodayOrderFills(&longbridge.GetTodayOrderFillsReq{
		Symbol:  "",
		OrderID: 0,
	})
	if err != nil {
		log.Fatalf("Error getting today's order fills: %v", err)
	}
	for _, o := range os {
		log.Printf("%+v\n", o)
	}

	// Output:
}

func ExampleTradeClient_GetMarginRatio() {
	c, err := longbridge.NewTradeClient(&longbridge.Config{
		BaseURL:       "https://openapi.longbridgeapp.com",
		AccessToken:   os.Getenv("LB_ACCESS_TOKEN"),
		QuoteEndpoint: "tcp://openapi-quote.longbridgeapp.com:2020",
		AppKey:        os.Getenv("LB_APP_KEY"),
		AppSecret:     os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	mr, err := c.GetMarginRatio("700.HK")
	if err != nil {
		log.Fatalf("Error getting margin ratio: %v", err)
	}
	log.Printf("%#v", mr)

	// Output:
}
