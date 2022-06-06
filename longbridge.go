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

// Package longbridge implements the high level long bridge trading API.
// It includes two sets: one for RESTful JSON API, and the other for long connection based on protobuf.
//
// To use the API, create a longbridge.Client and call the RESTful APIs on the client, or call the
// methods of client.Trade or client.Quote to access the data through trade and quote long connection respectively.
//
// For debugging, set -logtostderr and -v=<log level> to print out the internal protocol flow to console.
// In the implementation, glog is used to print different level of logging information.
// The log level rule is:
// 	Without V(n), show the log that are essential to be known by caller.
// 	V(2) for normal verbose notification for error cases.
// 	V(3) for deubgging information and consists detailed variable information.
//	V(4) for high frequency repeated debugging information.
package longbridge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

const (
	BaseURL = "https://openapi.longbridgeapp.com"
	// Ref: https://open.longbridgeapp.com/docs/socket/protocol/handshake#websocket-链接如何握手
	defaultTradeWebSocketEndPoint = "wss://openapi-trade.longbridgeapp.com?version=1&codec=1&platform=9"
	defaultQuoteWebSocketEndPoint = "wss://openapi-quote.longbridgeapp.com?version=1&codec=1&platform=9"
)

// Config contains the network address and the necessary credential information for accessing long bridge service.
// We use a simple config struct instead of functional options pattern for simplicity.
type Config struct {
	BaseURL       string
	TradeEndpoint string
	QuoteEndpoint string
	AccessToken   string
	AppKey        string
	AppSecret     string
}

type Request struct {
	Cmd  uint32
	Body proto.Message
}

// Client is the interface to call long bridge trade and quote service. For normal trading APIs, it uses HTTP protocol
// to manipulate the orders, such as client.PlaceOrder, etc.
//
// For long connection, setting the end points to url starting with "tcp://" will use TCP connection, while "wss://" for
// secured web socket connection. If not set, it will choose secured web socket by default.
// The order notification is done by setting the Trade.OnChangeOrder handler after the trade connection is enabled. e.g.,
//	client.Trade.OnOrderChange = func(order *Order) {...}
//	client.Trade.Enable(true)
// For quote related APIs, call client.Quote.GetXXX to use the pull service after enabling the quote connection.
// For quote service notification, set client.Quote.OnXXX callback and call client.Quote.SubscribePush to
// subscribe the quote data notification service.
type Client struct {
	config  *Config
	limiter *rate.Limiter
	Quote   *QuoteLongConn
	Trade   *TradeLongConn
}

func NewClient(conf *Config) (*Client, error) {
	if conf == nil || conf.BaseURL == "" || conf.AccessToken == "" || conf.AppKey == "" || conf.AppSecret == "" {
		return nil, fmt.Errorf("invalid config with empty key inside: %#v", conf)
	}
	// Min interval is 0.02 second for HTTP API
	c := &Client{config: conf, limiter: rate.NewLimiter(rate.Every(20*time.Microsecond), 1)}
	qe := c.config.QuoteEndpoint
	if qe == "" {
		qe = defaultQuoteWebSocketEndPoint
	}
	te := c.config.TradeEndpoint
	if te == "" {
		te = defaultTradeWebSocketEndPoint
	}
	c.Quote = newQuoteLongConn(qe, c)
	c.Trade = newTradeLongConn(te, c)
	return c, nil
}

func (c *Client) Close() {
	c.Quote.Enable(false)
	c.Trade.Enable(false)
}

// getOTP fetches the OTP (One-Time Password) from web API end point to support long-polling connection to trade gateway.
func (c *Client) getOTP() (string, error) {
	defer trace("requesting OTP")()
	req, err := c.createSignedRequest(GET, getOTPURLPath, nil, nil, time.Now())
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling long bridge API: %v", err)
	}
	defer resp.Body.Close()
	var data struct {
		Code    int
		Message string
		Data    struct {
			OTP string
		}
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return "", fmt.Errorf("error decoding the JSON data in response body: %v", err)
	}
	if data.Code != 0 {
		return "", fmt.Errorf("invalid code for OTP request, data obtained %#v", data)
	}
	return data.Data.OTP, nil
}

type AccessTokenInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token       string    `json:"token"`
		ExpiredAt   time.Time `json:"expired_at"`
		IssuedAt    time.Time `json:"issued_at"`
		AccountInfo struct {
			MemberID       string `json:"member_id"`
			Aaid           string `json:"aaid"`
			AccountChannel string `json:"account_channel"`
		} `json:"account_info"`
	} `json:"data"`
}

func (c *Client) RefreshAccessToken(expire time.Time) (*AccessTokenInfo, error) {
	const iso8601Fmt = "2006-01-02T15:04:05.000Z0700"
	defer trace("refresh access token")()
	var token AccessTokenInfo
	p := &params{}
	p.Add("expired_at", expire.UTC().Format(iso8601Fmt))
	if err := c.request(GET, refreshToken, p.Values(), nil, &token); err != nil {
		return nil, err
	}
	if token.Code != 0 {
		return nil, fmt.Errorf("error refreshing token: %s", token.Message)
	}
	return &token, nil
}
