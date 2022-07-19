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

func ExampleQuoteClient_GetStaticInfo() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ss, err := c.GetStaticInfo([]longbridge.Symbol{"700.HK", "AAPL.US"})
	if err != nil {
		log.Fatalf("Error getting static info: %v", err)
	}
	for _, s := range ss {
		log.Printf("%#v\n", s)
	}
	// Output:
}

func ExampleQuoteClient_GetRealtimeQuote() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	qs, err := c.GetRealtimeQuote([]longbridge.Symbol{"5.HK", "MSFT.US"})
	if err != nil {
		log.Fatalf("Error getting real time quote: %v", err)
	}
	for _, q := range qs {
		log.Printf("%#v\n", q)
	}
	// Output:
}

func ExampleQuoteClient_GetRealtimeOptionQuote() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	qs, err := c.GetRealtimeOptionQuote([]longbridge.Symbol{"AAPL230317P160000.US"})
	if err != nil {
		log.Fatalf("Error getting real time option quote: %v", err)
	}
	for _, q := range qs {
		log.Printf("%#v\n", q)
	}
	// Output:
}

func ExampleQuoteClient_GetRealtimeWarrantQuote() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	qs, err := c.GetRealtimeWarrantQuote([]longbridge.Symbol{"21125.HK"})
	if err != nil {
		log.Fatalf("Error getting real time warrant quote: %v", err)
	}
	for _, q := range qs {
		log.Printf("%#v\n", q)
	}
	// Output:
}

func ExampleQuoteClient_GetOrderBookList() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ol, err := c.GetOrderBookList("5.HK")
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

func ExampleQuoteClient_GetBrokerQueue() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	bq, err := c.GetBrokerQueue("5.HK")
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

func ExampleQuoteClient_GetBrokerInfo() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	bs, err := c.GetBrokerInfo()
	if err != nil {
		log.Fatalf("Error getting broker queue for: %v", err)
	}
	for _, b := range bs {
		log.Printf("%#v", b)
	}
	// Output:
}

func ExampleQuoteClient_GetTickers() {
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
	// Output:
}

func ExampleQuoteClient_GetIntradayLines() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ls, err := c.GetIntradayLines("AAPL.US")
	if err != nil {
		log.Fatalf("Error getting tickers for: %v", err)
	}
	for _, l := range ls {
		log.Printf("%#v", l)
	}
	// Output:
}

func ExampleQuoteClient_GetKLines() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ks, err := c.GetKLines("AAPL.US", longbridge.KLine1M, 10, longbridge.AdjustNone)
	if err != nil {
		log.Fatalf("Error getting klines: %v", err)
	}
	for _, k := range ks {
		log.Printf("%#v", k)
	}
	// Output:
}

func ExampleQuoteClient_GetOptionExpiryDates() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ks, err := c.GetOptionExpiryDates("AAPL.US")
	if err != nil {
		log.Fatalf("Error getting option chain expiry dates: %v", err)
	}
	for _, k := range ks {
		log.Printf("%#v", k)
	}
	// Output:
}

func ExampleQuoteClient_GetOptionStrikePrices() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ks, err := c.GetOptionStrikePrices("AAPL.US", "20230120")
	if err != nil {
		log.Fatalf("Error getting option chain strike price: %v", err)
	}
	for _, k := range ks {
		log.Printf("%#v", k)
	}
	// Output:
}

func ExampleQuoteClient_GetWarrantIssuers() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	is, err := c.GetWarrantIssuers()
	if err != nil {
		log.Fatalf("Error getting warrant issuer list: %v", err)
	}
	for _, i := range is {
		log.Printf("%#v", i)
	}
	// Output:
}

func ExampleQuoteClient_SearchWarrants() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ws, err := c.SearchWarrants(&longbridge.WarrantFilter{
		Symbol:     "700.HK",
		Language:   longbridge.SimplifiedChinese,
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

func ExampleQuoteClient_GetMarketTradePeriods() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ms, err := c.GetMarketTradePeriods()
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

func ExampleQuoteClient_GetTradeDates() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
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
	ds, err := c.GetTradeDates("HK", "20220530", "20220630")
	if err != nil {
		log.Fatalf("Error getting trade date list: %v", err)
	}
	for _, d := range ds {
		log.Printf("%#v", d)
	}
	// Output:
}

func ExampleQuoteClient_GetIntradayCapFlows() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
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
	fs, err := c.GetIntradayCapFlows("700.HK")
	if err != nil {
		log.Fatalf("Error getting intraday capital flow: %v", err)
	}
	for _, f := range fs {
		log.Printf("%+v", f)
	}
	// Output:
}

func ExampleQuoteClient_GetIntradayCapFlowDistribution() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
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
	cd, err := c.GetIntradayCapFlowDistribution("700.HK")
	if err != nil {
		log.Fatalf("Error getting intraday capital flow distribution: %v", err)
	}
	log.Printf("%+v", cd)
	// Output:
}

func ExampleQuoteClient_GetFinanceIndices() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
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
	fs, err := c.GetFinanceIndices([]longbridge.Symbol{"700.HK", "5.HK"},
		[]longbridge.FinanceIndex{longbridge.IndexLastDone, longbridge.IndexBalancePoint, longbridge.IndexCapitalFlow, longbridge.IndexExpiryDate})
	if err != nil {
		log.Fatalf("Error getting intraday capital flow distribution: %v", err)
	}
	for _, f := range fs {
		log.Printf("%+v", f)
	}
	// Output:
}

func ExampleQuoteClient_GetSubscriptions() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	ss, err := c.GetSubscriptions()
	if err != nil {
		log.Fatalf("Error getting subscription list: %v", err)
	}
	for _, s := range ss {
		log.Printf("%#v", s)
	}
	// Output:
}

func ExampleQuoteClient_SubscribePush() {
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
	c.OnPushQuote = func(q *longbridge.PushQuote) {
		log.Printf("Got realtime quote: %#v", q)
	}

	c.OnPushBrokers = func(b *longbridge.PushBrokers) {
		log.Printf("Got broker list for %s, seq=%d", b.Symbol, b.Sequence)
		for _, bid := range b.Bid {
			log.Printf("%#v", bid)
		}
		for _, ask := range b.Ask {
			log.Printf("%#v", ask)
		}
	}

	c.OnPushOrderBook = func(b *longbridge.PushOrderBook) {
		log.Printf("Got order books for %s, seq=%d", b.Symbol, b.Sequence)
		for _, bid := range b.Bid {
			log.Printf("%#v", bid)
		}
		for _, ask := range b.Ask {
			log.Printf("%#v", ask)
		}
	}

	ss, err := c.SubscribePush([]longbridge.Symbol{"AAPL.US"},
		[]longbridge.SubscriptionType{longbridge.SubscriptionTicker,
			longbridge.SubscriptionRealtimeQuote,
			longbridge.SubscriptionOrderBook,
			longbridge.SubscriptionBrokerQueue}, true)
	if err != nil {
		log.Fatalf("Error subscribe quote push: %v", err)
	}
	for _, s := range ss {
		log.Printf("%#v", s)
	}

	time.Sleep(30 * time.Second)
	if err := c.UnsubscribePush(nil,
		[]longbridge.SubscriptionType{longbridge.SubscriptionTicker,
			longbridge.SubscriptionRealtimeQuote,
			longbridge.SubscriptionOrderBook,
			longbridge.SubscriptionBrokerQueue}, true); err != nil {
		log.Fatalf("Error unsubscribing all symbols' quote push: %v", err)
	}
	// Output:
}

func ExampleQuoteClient_GetWatchedGroups() {
	c, err := longbridge.NewQuoteClient(&longbridge.Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	groups, err := c.GetWatchedGroups()
	if err != nil {
		log.Fatalf("Error getting watched groups: %v", err)
	}
	for _, g := range groups {
		log.Printf("%v %v: %v", g.ID, g.Name, g.Securities)
	}

	// Output:
}
