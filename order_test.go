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
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrder(t *testing.T) {
	assert := assert.New(t)
	// 1. Place req
	assert.Equal(map[string]string{
		"order_type": "MO", "remark": "Test", "side": "Buy", "submitted_price": "50", "submitted_quantity": "2", "symbol": "A.US", "time_in_force": "Day",
	}, (&PlaceOrderReq{Symbol: "A.US", Side: Buy, OrderType: MarketOrder, Price: 50, Quantity: 2, TimeInForce: DayOrder, Remark: "Test", ExpireDate: time.Time{}}).payload())

	f1 := 1.51
	f2 := 50.0100
	assert.Equal(map[string]string{
		"expire_date": "2022-06-01", "limit_offset": "50.01", "order_type": "ELO", "outside_rth": "ANY_TIME", "side": "Sell", "submitted_quantity": "100", "symbol": "15.HK", "time_in_force": "GTD", "trailing_amount": "1.51", "trailing_percent": "50.01", "trigger_price": "1.51",
	}, (&PlaceOrderReq{Symbol: "15.HK", OrderType: EnhancedLimitOrder, Quantity: 100, TriggerPrice: f1, LimitOffset: f2, TrailingAmount: f1, TrailingPercent: f2, ExpireDate: time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC), Side: Sell, OutsideRTH: AnyTime, TimeInForce: GoodTilDate}).payload())

	var resp placeOrderResp
	assert.Nil(decodeResp(bytes.NewReader([]byte(`{"code": 0,"message": "success", "data": {
		"order_id": "683615454870679600"
	}}`)), &resp))
	assert.EqualValues("683615454870679600", resp.Data.OrderID)

	// 2. Modify order
	assert.Equal(map[string]string{"order_id": "1", "quantity": "100"},
		(&ModifyOrderReq{OrderID: "1", Quantity: 100}).payload())
	assert.Equal(map[string]string{"limit_offset": "3.2", "order_id": "1", "price": "1.1", "quantity": "100", "remark": "...", "trailing_amount": "4", "trailing_percent": "5", "trigger_price": "2.5"},
		(&ModifyOrderReq{OrderID: "1", Quantity: 100, Price: 1.1, TriggerPrice: 2.5, LimitOffset: 3.2, TrailingAmount: 4, TrailingPercent: 5, Remark: "..."}).payload())

	// 3. Get history order
	assert.Equal("", (&GetHistoryOrdersReq{}).params().Encode())
	assert.Equal("symbol=1.HK", (&GetHistoryOrdersReq{Symbol: "1.HK"}).params().Encode())
	assert.Equal("end_at=2000&market=US&side=Buy&start_at=1000&status=FilledStatus&status=NewStatus&status=PartialFilledStatus&status=PendingCancelStatus",
		(&GetHistoryOrdersReq{Status: []OrderStatus{FilledStatus, NewStatus, PartialFilledStatus, PendingCancelStatus}, Side: Buy, Market: US, StartTimestamp: 1000, EndTimestamp: 2000}).params().Encode())

	var horderResp histroyOrderResp
	assert.Nil(decodeResp(strings.NewReader(`{
  "code": 0,
  "message": "success",
  "data": {
    "has_more": true,
    "orders": [
      {
        "currency": "HKD",
        "executed_price": "10.000",
        "executed_quantity": "2",
        "expire_date": "2022-06-10",
        "last_done": "?",
        "limit_offset": "5",
        "msg": "message",
        "order_id": "706388312699592704",
        "order_type": "ELO",
        "outside_rth": "UnknownOutsideRth",
        "price": "11.900",
        "quantity": "200",
        "side": "Buy",
        "status": "RejectedStatus",
        "stock_name": "Bank of East Asia Ltd/The",
        "submitted_at": "1651644897",
        "symbol": "23.HK",
        "tag": "Normal",
        "time_in_force": "Day",
        "trailing_amount": "50",
        "trailing_percent": "5",
        "trigger_at": "1",
        "trigger_price": "2.5",
        "trigger_status": "NOT_USED",
        "updated_at": "1651644898"
      }
    ]
  }
}`), &horderResp))
	order, err := horderResp.Data.Orders[0].toOrder()
	assert.Nil(err)
	assert.True(horderResp.Data.HasMore)
	assert.Equal(&Order{
		Currency:         "HKD",
		ExecutedPrice:    10,
		ExecutedQuantity: 2,
		ExpireDate:       "2022-06-10", LastDone: "?",
		LimitOffset: 5,
		Msg:         "message",
		OrderID:     706388312699592704,
		OrderType:   "ELO", OutsideRTH: "UnknownOutsideRth",
		Price:    11.9,
		Quantity: 200, Side: "Buy", Status: "RejectedStatus",
		StockName:          "Bank of East Asia Ltd/The",
		SubmittedTimestamp: 1651644897,
		Symbol:             "23.HK", Tag: "Normal",
		TimeInForce:      "Day",
		TrailingAmount:   50,
		TrailingPercent:  5,
		TriggerTimestamp: 1,
		TriggerPrice:     2.5,
		TriggerStatus:    "NOT_USED", UpdatedTimestamp: 1651644898,
	}, order)
}

func TestOrderFill(t *testing.T) {
	assert := assert.New(t)
	var resp orderFillResp
	assert.Nil(decodeResp(strings.NewReader(`{
  "code": 0,
  "message": "success",
  "data": {
    "has_more": true,
    "trades": [
      {
        "order_id": "693664675163312128",
        "price": "388",
        "quantity": "100",
        "symbol": "700.HK",
        "trade_done_at": "1648611351",
        "trade_id": "693664675163312128-1648611351433741210"
      }
    ]
  }
}`), &resp))
	fill, err := resp.Data.Trades[0].toOrderFill()
	assert.Nil(err)
	assert.True(resp.Data.HasMore)
	assert.Equal(&OrderFill{OrderID: 693664675163312128, Price: 388, Quantity: 100, Symbol: "700.HK", TradeDoneTimestamp: 1648611351, TradeID: "693664675163312128-1648611351433741210"}, fill)
}
