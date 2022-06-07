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
	"encoding/json"
	"time"

	"github.com/deepln-io/longbridge-goapi/internal/pb/trade"
	"github.com/deepln-io/longbridge-goapi/internal/protocol"
	"google.golang.org/protobuf/proto"

	"github.com/golang/glog"
)

// tradeLongConn handles the long connection for trade related notification. Currently it supports order change
// notification only. To receive the order change notification, set the callback OnOrderChange and call Enable(true).
type tradeLongConn struct {
	*longConn
	OnOrderChange func(order *Order)
}

func newTradeLongConn(endpoint string, provider otpProvider) *tradeLongConn {
	c := &tradeLongConn{longConn: newLongConn(endpoint, provider)}
	c.onPush = c.handlePushPkg
	return c
}

func (c *tradeLongConn) Enable(enable bool) {
	if enable && c.cancel == nil {
		go c.Subscribe([]string{"private"})
	}
	c.longConn.Enable(enable)
}

// Subscribe subscribes the topic for push notification. Currently for trade API only "private" has effect.
func (c *tradeLongConn) Subscribe(topics []string) error {
	header := c.getReqHeader(protocol.CmdSub)
	var resp trade.SubResponse
	if err := c.Call("subscription", header, &trade.Sub{Topics: topics}, &resp, 10*time.Second); err != nil {
		return err
	}

	success := make(map[string]bool)
	for _, topic := range resp.Success {
		success[topic] = true
	}
	fail := make(map[string]*trade.SubResponse_Fail)
	for _, f := range resp.Fail {
		fail[f.Topic] = f
	}
	current := make(map[string]bool)
	for _, topic := range resp.Current {
		current[topic] = true
	}
	var errs joinErrors
	for _, topic := range topics {
		if f, ok := fail[topic]; ok {
			errs.Add(f.Topic, f.Reason)
		} else if !success[topic] && !current[topic] {
			errs.Add(topic, "topic ignored")
		}
	}
	return errs.ToError()
}

func (c *tradeLongConn) Unsubscribe(topics []string) error {
	header := c.getReqHeader(protocol.CmdUnSub)
	var resp trade.UnsubResponse
	if err := c.Call("unsubscription", header, &trade.Unsub{Topics: topics}, &resp, 10*time.Second); err != nil {
		return err
	}

	current := make(map[string]bool)
	for _, topic := range resp.Current {
		current[topic] = true
	}
	var errs joinErrors
	for _, topic := range topics {
		if current[topic] {
			errs.Add(topic, "topic is still in current subscription")
		}
	}
	return errs.ToError()
}

// https://open.longbridgeapp.com/docs/trade/trade-definition#websocket-推送通知
type orderChange struct {
	Event string `json:"event"`
	Data  struct {
		Side             string `json:"side"`
		StockName        string `json:"stock_name"`
		Quantity         string `json:"quantity"`
		Symbol           Symbol `json:"symbol"`
		OrderType        string `json:"order_type"`
		Price            string `json:"price"`
		ExecutedQuantity string `json:"executed_quantity"`
		ExecutedPrice    string `json:"executed_price"`
		OrderID          string `json:"order_id"`
		Currency         string `json:"currency"`
		Status           string `json:"status"`
		SubmittedAt      string `json:"submitted_at"`
		UpdatedAt        string `json:"updated_at"`
		TriggerPrice     string `json:"trigger_price"`
		Msg              string `json:"msg"`
		Tag              string `json:"tag"`
		TriggerStatus    string `json:"trigger_status"`
		TriggerAt        string `json:"trigger_at"`
		TrailingAmount   string `json:"trailing_amount"`
		TrailingPercent  string `json:"trailing_percent"`
		LimitOffset      string `json:"limit_offset"`
		AccountNo        string `json:"account_no"`
	} `json:"data"`
}

func (oc *orderChange) toOrder() (*Order, error) {
	p := &parser{}
	order := &Order{
		Currency:      oc.Data.Currency,
		Msg:           oc.Data.Msg,
		OrderType:     OrderType(oc.Data.OrderType),
		OutsideRTH:    RTHOnly,
		Side:          TrdSide(oc.Data.Side),
		Status:        OrderStatus(oc.Data.Status),
		StockName:     oc.Data.StockName,
		Symbol:        oc.Data.Symbol,
		Tag:           OrderTag(oc.Data.Tag),
		TriggerStatus: TriggerStatus(oc.Data.TriggerStatus),
	}
	p.parse("executed_price", oc.Data.ExecutedPrice, &order.ExecutedPrice)
	p.parse("executed_quantity", oc.Data.ExecutedQuantity, &order.ExecutedQuantity)
	p.parse("limit_offset", oc.Data.LimitOffset, &order.LimitOffset)
	p.parse("order_id", oc.Data.OrderID, &order.OrderID)
	p.parse("price", oc.Data.Price, &order.Price)
	p.parse("quantity", oc.Data.Quantity, &order.Quantity)
	p.parse("submitted_at", oc.Data.SubmittedAt, &order.SubmittedTimestamp)
	p.parse("trailing_amount", oc.Data.TrailingAmount, &order.TrailingAmount)
	p.parse("trailing_percentage", oc.Data.TrailingPercent, &order.TrailingPercent)
	p.parse("trigger_price", oc.Data.TriggerPrice, &order.TriggerPrice)
	p.parse("updated_at", oc.Data.UpdatedAt, &order.UpdatedTimestamp)
	if err := p.Error(); err != nil {
		return nil, err
	}
	return order, nil
}

func (c *tradeLongConn) handlePushPkg(header *protocol.PushPkgHeader, body []byte, pkgErr error) {
	var resp trade.Notification
	if err := proto.Unmarshal(body, &resp); err != nil {
		glog.V(2).Infof("Discard ill formatted notification package: %v", err)
		return
	}
	if c.OnOrderChange != nil {
		if pushContentType(resp.ContentType) != pushContentTypeJSON {
			glog.V(2).Infof("Discard unsupported notification with content type %v, long bridge supports JSON only", resp.ContentType)
			return
		}
		var oc orderChange
		if err := json.Unmarshal(resp.Data, &oc); err != nil {
			glog.V(2).Infof("Error decoding the order change notification, content type=%v content len=%d: %v",
				resp.ContentType, len(resp.Data), err)
			return
		}
		glog.V(3).Infof("Order event: %s", oc.Event)
		order, err := oc.toOrder()
		if err != nil {
			glog.V(2).Infof("Error converting order change %#v to order: %v", oc, err)
			return
		}
		c.OnOrderChange(order)
	}
}

type TradeClient struct {
	*client
	*tradeLongConn
}

func NewTradeClient(conf *Config) (*TradeClient, error) {
	c, err := newClient(conf)
	if err != nil {
		return nil, err
	}
	tc := newTradeLongConn(c.config.TradeEndpoint, c)
	return &TradeClient{client: c, tradeLongConn: tc}, nil
}

func (c *TradeClient) Close() {
	c.tradeLongConn.Enable(false)
	c.client.Close()
}
