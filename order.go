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
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/golang/glog"
)

type OrderType string
type TrdSide string
type TimeInForce string
type OutsideRTH string // Outside regular trading hours
type OrderStatus string
type Market string
type OrderTag string
type TriggerStatus string

const (
	orderURLPath            urlPath = version + "/trade/order"
	historyOrderURLPath     urlPath = orderURLPath + "/history"
	todayOrderURLPath       urlPath = orderURLPath + "/today"
	orderFillURLPath        urlPath = version + "/trade/execution"
	todayOrderFillURLPath   urlPath = orderFillURLPath + "/today"
	historyOrderFillURLPath urlPath = orderFillURLPath + "/history"

	LimitOrder           OrderType = "LO"
	EnhancedLimitOrder   OrderType = "ELO"
	MarketOrder          OrderType = "MO"
	AtAuctionMarketOrder OrderType = "AO"
	AtAuctionLimitOrder  OrderType = "ALO"
	OddLotsOrder         OrderType = "ODD"     // 碎股單掛單
	LimitIfTouched       OrderType = "LIT"     // 觸價限價單
	MarketIfTouched      OrderType = "MIT"     // 觸價市價單
	TSLPAMT              OrderType = "TSLPAMT" // Trailing Limit If Touched (Trailing Amount) 跟蹤止損限價單 (跟蹤金额)
	TSLPPCT              OrderType = "TSLPPCT" // Trailing Limit If Touched (Trailing Percent) 跟蹤止損限價單 (跟蹤漲跌幅)
	TSMAMT               OrderType = "TSMAMT"  // Trailing Market If Touched (Trailing Amount) 跟蹤止損市價單 (跟蹤金额)
	TSMPCT               OrderType = "TSMPCT"  // Trailing Market If Touched (Trailing Percent) 跟蹤止損市價單 (跟蹤漲跌幅)

	Buy  TrdSide = "Buy"
	Sell TrdSide = "Sell"

	DayOrder      TimeInForce = "Day" // 當日有效
	GoodTilCancel TimeInForce = "GTC" // 撤單前有效
	GoodTilDate   TimeInForce = "GTD" // 到期前有效

	RTHOnly OutsideRTH = "RTH_ONLY" // Regular trading hour only
	AnyTime OutsideRTH = "ANY_TIME"

	NotReported          OrderStatus = "NotReported"          // 待提交
	ReplacedNotReported  OrderStatus = "ReplacedNotReported"  // 待提交 (改單成功)
	ProtectedNotReported OrderStatus = "ProtectedNotReported" // 待提交 (保價訂單)
	VarietiesNotReported OrderStatus = "VarietiesNotReported" // 待提交 (條件單)
	FilledStatus         OrderStatus = "FilledStatus"         // 已成交
	WaitToNew            OrderStatus = "WaitToNew"            // 已提待報
	NewStatus            OrderStatus = "NewStatus"            // 已委托
	WaitToReplace        OrderStatus = "WaitToReplace"        // 修改待報
	PendingReplaceStatus OrderStatus = "PendingReplaceStatus" // 待修改
	ReplacedStatus       OrderStatus = "ReplacedStatus"       // 已修改
	PartialFilledStatus  OrderStatus = "PartialFilledStatus"  // 部分成交
	WaitToCancel         OrderStatus = "WaitToCancel"         // 撤銷待報
	PendingCancelStatus  OrderStatus = "PendingCancelStatus"  // 待撤回
	RejectedStatus       OrderStatus = "RejectedStatus"       // 已拒絕
	CanceledStatus       OrderStatus = "CanceledStatus"       // 已撤單
	ExpiredStatus        OrderStatus = "ExpiredStatus"        // 已過期
	PartialWithdrawal    OrderStatus = "PartialWithdrawal"    // 部分撤單

	NormalOrder OrderTag = "Normal" // Normal order
	GTCOrder    OrderTag = "GTC"    // Long term order
	GreyOrder   OrderTag = "Grey"   // Grey order 暗盤單

	NotUsed  TriggerStatus = "NOT_USED" // 未激活
	Deactive TriggerStatus = "DEACTIVE"
	Active   TriggerStatus = "ACTIVE"
	Released TriggerStatus = "RELEASED" // 已觸發
)

// PlaceOrderReq is a request to place order. Fields are optional unless marked as 'Required'.
type PlaceOrderReq struct {
	// Stock symbol: Required, use ticker.region format, example: AAPL.US
	Symbol Symbol
	// Order type: Required
	OrderType OrderType
	Price     float64 // For limit order (LO, ELO, ALO)
	// Quantity: Required
	Quantity        uint64
	TriggerPrice    float64
	LimitOffset     float64
	TrailingAmount  float64
	TrailingPercent float64
	// Expire date (in timezone HKT or EDT) for long term order
	ExpireDate time.Time
	// Trade side: Required
	Side       TrdSide
	OutsideRTH OutsideRTH
	// Time in force: Required
	TimeInForce TimeInForce
	Remark      string // Max 64 characters
}

func (r *PlaceOrderReq) payload() map[string]string {
	p := params{
		"symbol":             string(r.Symbol),
		"order_type":         string(r.OrderType),
		"submitted_quantity": strconv.FormatUint(r.Quantity, 10),
		"side":               string(r.Side),
		"time_in_force":      string(r.TimeInForce),
	}
	p.AddOptFloat("submitted_price", r.Price)
	p.AddOptFloat("trigger_price", r.TriggerPrice)
	p.AddOptFloat("limit_offset", r.LimitOffset)
	p.AddOptFloat("trailing_amount", r.TrailingAmount)
	p.AddOptFloat("trailing_percent", r.TrailingPercent)
	p.AddDate("expire_date", r.ExpireDate)
	p.Add("outside_rth", string(r.OutsideRTH))
	p.Add("remark", r.Remark)
	return p
}

type placeOrderResp struct {
	statusResp
	Data struct {
		OrderID uint64 `json:"order_id"`
	}
}

// PlaceOrder places an order. It returns order ID and error (if any).
func (c *Client) PlaceOrder(r *PlaceOrderReq) (uint64, error) {
	payload := r.payload()
	pdata, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	var resp placeOrderResp
	if err := c.request(POST, orderURLPath, nil, pdata, &resp); err != nil {
		return 0, err
	}
	if err := resp.CheckSuccess(); err != nil {
		return 0, err
	}
	return resp.Data.OrderID, nil
}

// ModifyOrderReq is a request to modify order.
type ModifyOrderReq struct {
	OrderID  uint64 // Required
	Quantity uint64 // Required
	// Price is optional, required for order of type LO/ELO/ALO/ODD/LIT
	Price           float64
	TriggerPrice    float64
	LimitOffset     float64
	TrailingAmount  float64
	TrailingPercent float64
	Remark          string // Max 64 characters
}

func (r *ModifyOrderReq) payload() map[string]string {
	p := params{
		"order_id": strconv.FormatUint(r.OrderID, 10),
		"quantity": strconv.FormatUint(r.Quantity, 10),
	}
	p.AddOptFloat("price", r.Price)
	p.AddOptFloat("trigger_price", r.TriggerPrice)
	p.AddOptFloat("limit_offset", r.LimitOffset)
	p.AddOptFloat("trailing_amount", r.TrailingAmount)
	p.AddOptFloat("trailing_percent", r.TrailingPercent)
	p.Add("remark", r.Remark)
	return p
}

// ModifyOrder modifies an order. Order ID and quantity are required.
func (c *Client) ModifyOrder(r *ModifyOrderReq) error {
	payload := r.payload()
	pdata, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var resp statusResp
	if err := c.request(PUT, orderURLPath, nil, pdata, &resp); err != nil {
		return err
	}
	return resp.CheckSuccess()
}

// CancelOrder cancels an open order.
func (c *Client) CancelOrder(orderID uint64) error {
	var resp statusResp
	if err := c.request(DELETE, orderURLPath, map[string][]string{
		"order_id": {strconv.FormatUint(orderID, 10)},
	}, nil, &resp); err != nil {
		return err
	}
	return resp.CheckSuccess()
}

// GetOrdersReq is the request to get history/today orders. All fields are optional.
type GetOrdersReq struct {
	Symbol  Symbol
	Status  []OrderStatus
	Side    TrdSide
	Market  Market
	OrderID uint64 // For today order only
	// StartTimestamp and EndTimestamp are for histroy orders only
	StartTimestamp int64
	EndTimestamp   int64
	// IsAll indicates if getting all histroy orders.
	IsAll bool
}

type histroyOrderResp struct {
	statusResp
	Data struct {
		HasMore bool `json:"has_more"` // If has more records
		Orders  []*order
	}
}

type order struct {
	Currency         string
	ExecutedPrice    string `json:"executed_price"`
	ExecutedQuantity string `json:"executed_quantity"`
	ExpireDate       string `json:"expire_date"` // In format YYYY-MM-DD
	LastDone         string `json:"last_done"`
	LimitOffset      string `json:"limit_offset"`
	Msg              string
	OrderID          string `json:"order_id"`
	OrderType        string `json:"order_type"`
	OutsideRTH       string `json:"outside_rth"`
	Price            string
	Quantity         string
	Side             string
	Status           string
	StockName        string `json:"stock_name"`
	SubmittedAt      string `json:"submitted_at"` // UNIX timestamp
	Symbol           Symbol
	Tag              string
	TimeInForce      string `json:"time_in_force"`
	TrailingAmount   string `json:"trailing_amount"`
	TrailingPercent  string `json:"trailing_percent"`
	TriggerAt        string `json:"trigger_at"`
	TriggerPrice     string `json:"trigger_price"`
	TriggerStatus    string `json:"trigger_status"`
	UpdatedAt        string `json:"updated_at"`
}

func (o *order) ToOrder() (*Order, error) {
	order := &Order{
		Currency:         o.Currency,
		ExecutedQuantity: 0,
		ExpireDate:       o.ExpireDate,
		LastDone:         o.LastDone,
		Msg:              o.Msg,
		OrderType:        OrderType(o.OrderType),
		OutsideRTH:       OutsideRTH(o.OutsideRTH),
		Side:             TrdSide(o.Side),
		Status:           OrderStatus(o.Status),
		StockName:        o.StockName,
		Symbol:           o.Symbol,
		Tag:              OrderTag(o.Tag),
		TimeInForce:      TimeInForce(o.TimeInForce),
		TriggerTimestamp: 0,
		TriggerStatus:    TriggerStatus(o.TriggerStatus),
		UpdatedTimestamp: 0,
	}
	p := &parser{}
	p.parse("order_id", o.OrderID, &order.OrderID)
	p.parse("quantity", o.Quantity, &order.Quantity)
	p.parse("submitted_at", o.SubmittedAt, &order.SubmittedTimestamp)
	p.parse("executed_price", o.ExecutedPrice, &order.ExecutedPrice)
	p.parse("executed_quantity", o.ExecutedQuantity, &order.ExecutedQuantity)
	p.parse("limit_offset", o.LimitOffset, &order.LimitOffset)
	p.parse("price", o.Price, &order.Price)
	p.parse("trailing_amount", o.TrailingAmount, &order.TrailingAmount)
	p.parse("trailing_percentage", o.TrailingPercent, &order.TrailingPercent)
	p.parse("trigger_at", o.TriggerAt, &order.TriggerTimestamp)
	p.parse("trigger_price", o.TriggerPrice, &order.TriggerPrice)
	p.parse("updated_at", o.UpdatedAt, &order.UpdatedTimestamp)
	if err := p.Error(); err != nil {
		return nil, err
	}
	return order, nil
}

type Order struct {
	Currency           string // Required
	ExecutedPrice      float64
	ExecutedQuantity   uint64
	ExpireDate         string // In format of 'YYYY-MM-DD' if provided
	LastDone           string
	LimitOffset        float64
	Msg                string
	OrderID            uint64    // Required
	OrderType          OrderType // Required
	OutsideRTH         OutsideRTH
	Price              float64
	Quantity           uint64      // Required
	Side               TrdSide     // Required
	Status             OrderStatus // Required
	StockName          string      // Required
	SubmittedTimestamp int64       // Required
	Symbol             Symbol      // Required
	Tag                OrderTag    // Required
	TimeInForce        TimeInForce // Required
	TrailingAmount     float64
	TrailingPercent    float64
	TriggerTimestamp   int64
	TriggerPrice       float64
	TriggerStatus      TriggerStatus
	UpdatedTimestamp   int64
}

func (r *GetOrdersReq) params() url.Values {
	p := &params{}
	p.Add("symbol", string(r.Symbol))
	p.Add("side", string(r.Side))
	p.Add("market", string(r.Market))
	p.AddOptInt("start_at", r.StartTimestamp)
	p.AddOptInt("end_at", r.EndTimestamp)
	p.AddOptUint("order_id", r.OrderID)
	vals := p.Values()
	for _, s := range r.Status {
		vals.Add("status", string(s))
	}
	return vals
}

// GetHistoryOrders get history orders in ascending update time order.
func (c *Client) GetHistoryOrders(r *GetOrdersReq) ([]*Order, error) {
	if !r.IsAll {
		orders, _, err := c.getOrders(historyOrderURLPath, r)
		return orders, err
	}
	hasMore := true
	var orders []*Order
	orderIDs := map[uint64]bool{}
	for hasMore {
		batchOrders, more, err := c.getOrders(historyOrderURLPath, r)
		glog.V(3).Infof("Fetched histroy orders for request %+v [Has more: %t, Error: %v]", r, more, err)
		if err != nil {
			return nil, err
		}
		if len(batchOrders) == 0 {
			return orders, nil
		}
		for _, o := range batchOrders {
			if !orderIDs[o.OrderID] {
				orders = append(orders, o)
				orderIDs[o.OrderID] = true
			}
		}
		// Get next batch orders starts from last update time. End timestamp remains the same if provided
		r.StartTimestamp = batchOrders[len(batchOrders)-1].UpdatedTimestamp
		hasMore = more
	}
	return orders, nil
}

// GetTodayOrders get today orders in ascending update time order.
func (c *Client) GetTodayOrders(r *GetOrdersReq) ([]*Order, error) {
	orders, _, err := c.getOrders(todayOrderURLPath, r)
	return orders, err
}

func (c *Client) getOrders(url urlPath, r *GetOrdersReq) ([]*Order, bool, error) {
	var resp histroyOrderResp
	var orders []*Order
	if err := c.request(GET, url, r.params(), nil, &resp); err != nil {
		return nil, false, err
	}
	if err := resp.CheckSuccess(); err != nil {
		return nil, false, err
	}
	var errs joinErrors
	for _, o := range resp.Data.Orders {
		order, err := o.ToOrder()
		if err != nil {
			errs.Add(fmt.Sprintf("Order %+v", o), err)
			continue
		}
		orders = append(orders, order)
	}
	return orders, resp.Data.HasMore, errs.ToError()
}

// OrderFillReq is a request to get today or history order fills (executions). All fields are optional.
type OrderFillReq struct {
	Symbol Symbol
	// OrderID is for today order fills only
	OrderID uint64
	// Start and end timestamp are for history order fills only
	StartTimestamp int64
	EndTimestamp   int64
}

type orderFillResp struct {
	statusResp
	Data struct {
		HasMore bool `json:"has_more"`
		Trades  []*orderFill
	}
}

type orderFill struct {
	OrderID     string `json:"order_id"`
	Price       string
	Quantity    string
	Symbol      Symbol
	TradeDoneAt string `json:"trade_done_at"`
	TradeID     string `json:"trade_id"`
}

func (f *orderFill) ToOrderFill() (*OrderFill, error) {
	parser := &parser{}
	fill := &OrderFill{
		Symbol: f.Symbol, TradeID: f.TradeID,
	}
	parser.parse("order_id", f.OrderID, &fill.OrderID)
	parser.parse("price", f.Price, &fill.Price)
	parser.parse("quantity", f.Quantity, &fill.Quantity)
	parser.parse("trade_done_at", f.TradeDoneAt, &fill.TradeDoneTimestamp)
	return fill, parser.errs.ToError()
}

type OrderFill struct {
	OrderID            uint64
	Price              float64
	Quantity           uint64
	Symbol             Symbol
	TradeDoneTimestamp int64
	TradeID            string
}

func (r *OrderFillReq) params() url.Values {
	p := params{}
	p.Add("symbol", string(r.Symbol))
	p.AddOptUint("order_id", r.OrderID)
	p.AddOptInt("start_at", r.StartTimestamp)
	p.AddOptInt("end_at", r.EndTimestamp)
	return p.Values()
}

// GetHistoryOrderFill get history order fills in ascending trade done time.
func (c *Client) GetHistoryOrderFill(r *OrderFillReq) ([]*OrderFill, error) {
	hasMore := true
	var orderFills []*OrderFill
	fillIDs := make(map[string]bool)
	for hasMore {
		fills, more, err := c.getOrderFill(historyOrderFillURLPath, r)
		glog.V(3).Infof("Fetched order fills for request %+v [Has more: %t, Error: %v]", r, more, err)
		if err != nil {
			return nil, err
		}
		if len(fills) == 0 {
			return orderFills, nil
		}
		hasMore = more
		for _, fill := range fills {
			if !fillIDs[fill.TradeID] {
				fillIDs[fill.TradeID] = true
				orderFills = append(orderFills, fill)
			}
		}
		// Fetch next batch at last trade done time
		r.StartTimestamp = fills[len(fills)-1].TradeDoneTimestamp
	}
	return orderFills, nil
}

// GetTodayOrderFill get today order fills in ascending trade done time.
func (c *Client) GetTodayOrderFill(r *OrderFillReq) ([]*OrderFill, error) {
	fills, _, err := c.getOrderFill(todayOrderFillURLPath, r)
	return fills, err
}

func (c *Client) getOrderFill(url urlPath, r *OrderFillReq) ([]*OrderFill, bool, error) {
	var resp orderFillResp
	if err := c.request(GET, url, r.params(), nil, &resp); err != nil {
		return nil, false, err
	}
	if err := resp.CheckSuccess(); err != nil {
		return nil, false, err
	}
	var fills []*OrderFill
	var errs joinErrors
	for _, f := range resp.Data.Trades {
		fill, err := f.ToOrderFill()
		if err != nil {
			errs.Add(fmt.Sprintf("OrderFill %+v", f), err)
			continue
		}
		fills = append(fills, fill)
	}
	return fills, resp.Data.HasMore, errs.ToError()
}
