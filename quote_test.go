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

package longbridge

import (
	"log"
	"os"
	"time"
)

func ExampleQuoteLongConn_GetStaticInfo() {
	c, err := NewClient(&Config{
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
	// Output:
}

func ExampleQuoteLongConn_GetRealtimeQuote() {
	c, err := NewClient(&Config{
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
	qs, err := c.Quote.GetRealtimeQuote([]string{"5.HK", "MSFT.US"})
	if err != nil {
		log.Fatalf("Error getting real time quote: %v", err)
	}
	for _, q := range qs {
		log.Printf("%#v\n", q)
	}
	// Output:
}

func ExampleQuoteLongConn_GetRealtimeOptionQuote() {
	c, err := NewClient(&Config{
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
	qs, err := c.Quote.GetRealtimeOptionQuote([]string{"AAPL230317P160000.US"})
	if err != nil {
		log.Fatalf("Error getting real time option quote: %v", err)
	}
	for _, q := range qs {
		log.Printf("%#v\n", q)
	}
	// Output:
}

func ExampleQuoteLongConn_GetRealtimeWarrantQuote() {
	c, err := NewClient(&Config{
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
	qs, err := c.Quote.GetRealtimeWarrantQuote([]string{"21125.HK"})
	if err != nil {
		log.Fatalf("Error getting real time warrant quote: %v", err)
	}
	for _, q := range qs {
		log.Printf("%#v\n", q)
	}
	// Output:
}

func ExampleQuoteLongConn_GetOrderBookList() {
	c, err := NewClient(&Config{
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
	ol, err := c.Quote.GetOrderBookList("5.HK")
	if err != nil {
		log.Fatalf("Error getting order book for: %v", err)
	}
	log.Printf("%s\n", ol.Symbol)
	for _, bid := range ol.Bid {
		log.Printf("%#v\n", bid)
	}
	for _, ask := range ol.Ask {
		log.Printf("%#v\n", ask)
	}
	// Output:
}

func ExampleQuoteLongConn_GetBrokerQueue() {
	c, err := NewClient(&Config{
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
	bq, err := c.Quote.GetBrokerQueue("5.HK")
	if err != nil {
		log.Fatalf("Error getting broker queue for: %v", err)
	}
	log.Printf("%s\n", bq.Symbol)
	for _, bid := range bq.Bid {
		log.Printf("%#v\n", bid)
	}
	for _, ask := range bq.Ask {
		log.Printf("%#v\n", ask)
	}
	// Output:
}

func ExampleQuoteLongConn_GetBrokerInfo() {
	c, err := NewClient(&Config{
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
	bs, err := c.Quote.GetBrokerInfo()
	if err != nil {
		log.Fatalf("Error getting broker queue for: %v", err)
	}
	for _, b := range bs {
		log.Printf("%#v", b)
	}
	// Output:
}

func ExampleQuoteLongConn_GetTickers() {
	c, err := NewClient(&Config{
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
	ts, err := c.Quote.GetTickers("AAPL.US", 10)
	if err != nil {
		log.Fatalf("Error getting tickers for: %v", err)
	}
	for _, t := range ts {
		log.Printf("%#v", t)
	}
	// Output:
}

func ExampleQuoteLongConn_GetIntradayLines() {
	c, err := NewClient(&Config{
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
	ls, err := c.Quote.GetIntradayLines("AAPL.US")
	if err != nil {
		log.Fatalf("Error getting tickers for: %v", err)
	}
	for _, l := range ls {
		log.Printf("%#v", l)
	}
	// Output:
}

func ExampleQuoteLongConn_GetKLines() {
	c, err := NewClient(&Config{
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
	ks, err := c.Quote.GetKLines("AAPL.US", KLine1M, 10, AdjustNone)
	if err != nil {
		log.Fatalf("Error getting klines: %v", err)
	}
	for _, k := range ks {
		log.Printf("%#v", k)
	}
	// Output:
}

func ExampleQuoteLongConn_GetOptionExpiryDates() {
	c, err := NewClient(&Config{
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
	ks, err := c.Quote.GetOptionExpiryDates("AAPL.US")
	if err != nil {
		log.Fatalf("Error getting option chain expiry dates: %v", err)
	}
	for _, k := range ks {
		log.Printf("%#v", k)
	}
	// Output:
}

func ExampleQuoteLongConn_GetOptionStrikePrices() {
	c, err := NewClient(&Config{
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
	ks, err := c.Quote.GetOptionStrikePrices("AAPL.US", "20230120")
	if err != nil {
		log.Fatalf("Error getting option chain strike price: %v", err)
	}
	for _, k := range ks {
		log.Printf("%#v", k)
	}
	// Output:
}

func ExampleQuoteLongConn_GetWarrantIssuers() {
	c, err := NewClient(&Config{
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
	is, err := c.Quote.GetWarrantIssuers()
	if err != nil {
		log.Fatalf("Error getting warrant issuer list: %v", err)
	}
	for _, i := range is {
		log.Printf("%#v", i)
	}
	// Output:
}

func ExampleQuoteLongConn_SearchWarrants() {
	c, err := NewClient(&Config{
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
	ws, err := c.Quote.SearchWarrants(&WarrantFilter{
		Symbol:     "700.HK",
		Language:   SimplifiedChinese,
		SortBy:     0,
		SortOrder:  1,
		SortOffset: 1,
		PageSize:   10,
	})
	if err != nil {
		log.Fatalf("Error searching warrants: %v", err)
	}
	for _, w := range ws {
		log.Printf("%#v", w)
	}
	// Output:
}

func ExampleQuoteLongConn_GetMarketTradePeriods() {
	c, err := NewClient(&Config{
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
	ms, err := c.Quote.GetMarketTradePeriods()
	if err != nil {
		log.Fatalf("Error getting market periods: %v", err)
	}
	for _, m := range ms {
		log.Printf("%#v", m.Market)
		for _, s := range m.TradeSessions {
			log.Printf("%#v", s)
		}
	}
	// Output:
}

func ExampleQuoteLongConn_GetTradeDates() {
	c, err := NewClient(&Config{
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
	c.Quote.Enable(true)
	ds, err := c.Quote.GetTradeDates("HK", "20220530", "20220630")
	if err != nil {
		log.Fatalf("Error getting trade date list: %v", err)
	}
	for _, d := range ds {
		log.Printf("%#v", d)
	}
	// Output:
}

func ExampleQuoteLongConn_GetSubscriptions() {
	c, err := NewClient(&Config{
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
	ss, err := c.Quote.GetSubscriptions()
	if err != nil {
		log.Fatalf("Error getting subscription list: %v", err)
	}
	for _, s := range ss {
		log.Printf("%#v", s)
	}
	// Output:
}

func ExampleQuoteLongConn_SubscribePush() {
	c, err := NewClient(&Config{
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
	c.Quote.OnPushTickers = func(t *PushTickers) {
		log.Printf("Got tickers for %s, seq=%d", t.Symbol, t.Sequence)
		for _, ticker := range t.Tickers {
			log.Printf("Ticker: %#v", ticker)
		}
	}
	c.Quote.OnPushQuote = func(q *PushQuote) {
		log.Printf("Got realtime quote: %#v", q)
	}

	c.Quote.OnPushBrokers = func(b *PushBrokers) {
		log.Printf("Got broker list for %s, seq=%d", b.Symbol, b.Sequence)
		for _, bid := range b.Bid {
			log.Printf("%#v", bid)
		}
		for _, ask := range b.Ask {
			log.Printf("%#v", ask)
		}
	}

	c.Quote.OnPushOrderBook = func(b *PushOrderBook) {
		log.Printf("Got order books for %s, seq=%d", b.Symbol, b.Sequence)
		for _, bid := range b.Bid {
			log.Printf("%#v", bid)
		}
		for _, ask := range b.Ask {
			log.Printf("%#v", ask)
		}
	}

	ss, err := c.Quote.SubscribePush([]string{"AAPL.US"},
		[]SubscriptionType{SubscriptionTicker,
			SubscriptionRealtimeQuote,
			SubscriptionOrderBook,
			SubscriptionBrokerQueue}, true)
	if err != nil {
		log.Fatalf("Error subscribe quote push: %v", err)
	}
	for _, s := range ss {
		log.Printf("%#v", s)
	}

	time.Sleep(30 * time.Second)
	if err := c.Quote.UnsubscribePush(nil,
		[]SubscriptionType{SubscriptionTicker,
			SubscriptionRealtimeQuote,
			SubscriptionOrderBook,
			SubscriptionBrokerQueue}, true); err != nil {
		log.Fatalf("Error unsubscribing all symbols' quote push: %v", err)
	}
	// Output:
}
