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

// Package longbridge implements the long bridge trading and quote API.
// It provides two clients, TradeClient and QuoteClient, to access trading service and quote service respectively.
// The API implementation strictly follows the document at https://open.longbridgeapp.com/docs.
// For each exported API, we have related short test examples.
// To run the examples, set the environment variables for LB_APP_KEY, LB_APP_SECRET, and LB_ACCESS_TOKEN obtained
// from long bridge developer center page.
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
	defaultReconnectInterval = 3

	defaultBaseURL = "https://openapi.longbridgeapp.com"
	// Ref: https://open.longbridgeapp.com/docs/socket/protocol/handshake#websocket-链接如何握手
	defaultTradeWebSocketEndPoint = "wss://openapi-trade.longbridgeapp.com"
	defaultQuoteWebSocketEndPoint = "wss://openapi-quote.longbridgeapp.com"
)

// Config contains the network address and the necessary credential information for accessing long bridge service.
type Config struct {
	AccessToken       string // Required
	AppKey            string // Required
	AppSecret         string // Required
	BaseURL           string // Optional, if not set, use default base URL https://openapi.longbridgeapp.com
	TradeEndpoint     string // Optional, if not set, use wss://openapi-trade.longbridgeapp.com
	QuoteEndpoint     string // Optional, if not set, use wss://openapi-quote.longbridgeapp.com
	ReconnectInterval int    // Optional, if not set, use default 3 seconds for reconnecting to the above endpoints
}

type Request struct {
	Cmd  uint32
	Body proto.Message
}

type client struct {
	config   *Config
	limiters []*rate.Limiter
}

func newClient(conf *Config) (*client, error) {
	if conf == nil || conf.AccessToken == "" || conf.AppKey == "" || conf.AppSecret == "" {
		return nil, fmt.Errorf("invalid config with empty credential fields inside: %#v", conf)
	}
	cc := *conf
	if cc.BaseURL == "" {
		cc.BaseURL = defaultBaseURL
	}
	if cc.QuoteEndpoint == "" {
		cc.QuoteEndpoint = defaultQuoteWebSocketEndPoint
	}
	if cc.TradeEndpoint == "" {
		cc.TradeEndpoint = defaultTradeWebSocketEndPoint
	}
	if cc.ReconnectInterval <= 0 {
		cc.ReconnectInterval = defaultReconnectInterval
	}
	return &client{config: &cc,
			// Ref: https://open.longbridgeapp.com/docs/#使用权限及限制,
			// max 30 request per 30 seconds, and two requests need at least 0.02 seconds.
			limiters: []*rate.Limiter{
				rate.NewLimiter(rate.Every(20*time.Millisecond), 1),
				rate.NewLimiter(rate.Every(30*time.Second), 30)}},
		nil
}

func (c *client) Close() {
}

// getOTP fetches the OTP (One-Time Password) from web API end point to support long-polling connection to trade gateway.
func (c *client) getOTP() (string, error) {
	defer trace("requesting OTP")()
	req, err := c.createSignedRequest(httpGET, getOTPURLPath, nil, nil, time.Now())
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

func (c *client) RefreshAccessToken(expire time.Time) (*AccessTokenInfo, error) {
	const iso8601Fmt = "2006-01-02T15:04:05.000Z0700"
	defer trace("refresh access token")()
	var token AccessTokenInfo
	p := &params{}
	p.Add("expired_at", expire.UTC().Format(iso8601Fmt))
	if err := c.request(httpGET, refreshToken, p.Values(), nil, &token); err != nil {
		return nil, err
	}
	if token.Code != 0 {
		return nil, fmt.Errorf("error refreshing token: %s", token.Message)
	}
	return &token, nil
}
