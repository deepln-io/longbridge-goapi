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
	"fmt"
	"net/url"

	"github.com/golang/glog"
)

const (
	accountBalanceURLPath urlPath = "/v1/asset/account"
	cashflowURLPath       urlPath = "/v1/asset/cashflow"
	fundPositionURLPath   urlPath = "/v1/asset/fund"
	marginRatioURLPath    urlPath = "/v1/risk/margin-ratio"
	stockPositionURLPath  urlPath = "/v1/asset/stock"
)

type RiskLevel int

const (
	Low RiskLevel = iota + 1
	LowToMedium
	Medium
	MediumToHigh
	High
)

type AssetType int

const (
	CashAsset AssetType = iota + 1
	StockAsset
	FundAsset
)

type CashflowDirection int

const (
	Outflow CashflowDirection = iota + 1
	Inflow
)

type Cash struct {
	Withdraw  float64
	Available float64
	Frozen    float64
	Settling  float64
	Currency  string
}

type Balance struct {
	TotalCash           float64
	MaxFinanceAmount    float64
	RemainFinanceAmount float64
	RiskLevel           RiskLevel
	MarginCall          float64
	NetAssets           float64
	InitMargin          float64
	MaintenanceMargin   float64
	Currency            string
	Cashes              []*Cash
}

type balancesResp struct {
	statusResp
	Data struct {
		List []balance
	}
}

type balance struct {
	TotalCash           string `json:"total_cash"`
	MaxFinanceAmount    string `json:"max_finance_amount"`
	RemainFinanceAmount string `json:"remaining_finance_amount"`
	RiskLevel           string `json:"risk_level"`
	MarginCall          string `json:"margin_call"`
	NetAssets           string `json:"net_assets"`
	InitMargin          string `json:"init_margin"`
	MaintenanceMargin   string `json:"maintenance_margin"`
	Currency            string
	Cashes              []struct {
		Withdraw  string `json:"withdraw_cash"`
		Available string `json:"available_cash"`
		Frozen    string `json:"frozen_cash"`
		Settling  string `json:"settling_cash"`
		Currency  string
	} `json:"cash_infos"`
}

func (b *balance) toBalance() (*Balance, error) {
	p := &parser{}
	bal := &Balance{
		Currency: b.Currency,
	}
	p.parse("total_cash", b.TotalCash, &bal.TotalCash)
	p.parse("max_finance_amount", b.MaxFinanceAmount, &bal.MaxFinanceAmount)
	p.parse("remaining_finance_amount", b.RemainFinanceAmount, &bal.RemainFinanceAmount)
	var riskLevel int64
	p.parse("risk_level", b.RiskLevel, &riskLevel)
	bal.RiskLevel = RiskLevel(riskLevel)
	p.parse("margin_call", b.MarginCall, &bal.MarginCall)
	p.parse("net_assets", b.NetAssets, &bal.NetAssets)
	p.parse("init_margin", b.InitMargin, &bal.InitMargin)
	p.parse("maintenance_margin", b.MaintenanceMargin, &bal.MaintenanceMargin)
	for _, c := range b.Cashes {
		cash := &Cash{Currency: c.Currency}
		p.parse("withdraw_cash", c.Withdraw, &cash.Withdraw)
		p.parse("available_cash", c.Available, &cash.Available)
		p.parse("frozen_cash", c.Frozen, &cash.Frozen)
		p.parse("settling_cash", c.Settling, &cash.Settling)
		bal.Cashes = append(bal.Cashes, cash)
	}
	return bal, p.errs.ToError()
}

// GetAccountBalances get account balances.
func (c *TradeClient) GetAccountBalances() ([]*Balance, error) {
	var resp balancesResp
	if err := c.request(httpGET, accountBalanceURLPath, nil, nil, &resp); err != nil {
		return nil, err
	}
	return getBalances(&resp)
}

func getBalances(resp *balancesResp) ([]*Balance, error) {
	if err := resp.CheckSuccess(); err != nil {
		return nil, err
	}
	var bals []*Balance
	var errs joinErrors
	for _, b := range resp.Data.List {
		bal, err := b.toBalance()
		if err != nil {
			errs.Add(fmt.Sprintf("balance %+v", b), bal)
			continue
		}
		bals = append(bals, bal)
	}
	return bals, errs.ToError()
}

// CashflowReq is the request to get cash flows. Only field StartTimestamp and EndTimestamp are required.
type CashflowReq struct {
	StartTimestamp int64
	EndTimestamp   int64
	BusinessType   AssetType
	Symbol         Symbol
	// Page is start page (>= 1), default is 1.
	Page uint64
	// Size is page size (1 ~ 10000), default is 50.
	Size uint64
	// IsAll indicates if get all cash flows. Default is get cash flows in current page.
	IsAll bool
}

func (r *CashflowReq) params() url.Values {
	p := params{}
	p.AddInt("start_time", r.StartTimestamp)
	p.AddInt("end_time", r.EndTimestamp)
	p.AddOptInt("business_type", int64(r.BusinessType))
	p.Add("symbol", string(r.Symbol))
	p.AddOptUint("page", r.Page)
	p.AddOptUint("size", r.Size)
	return p.Values()
}

type Cashflow struct {
	TransactionFlowName string
	Direction           CashflowDirection
	BusinessType        AssetType
	Balance             float64
	Currency            string
	BusinessTimestamp   int64
	Symbol              Symbol // Optional
	Description         string // Optional
}

type cashflow struct {
	TransactionFlowName string `json:"transaction_flow_name"`
	Direction           int
	BusinessType        int `json:"business_type"`
	Balance             string
	Currency            string
	BusinessTime        string `json:"business_time"`
	Symbol              Symbol
	Description         string
}

func (cf *cashflow) ToCashflow() (*Cashflow, error) {
	cashflow := &Cashflow{
		TransactionFlowName: cf.TransactionFlowName,
		Currency:            cf.Currency,
		Description:         cf.Description,
		Symbol:              cf.Symbol,
	}
	p := &parser{}
	cashflow.Direction = CashflowDirection(cf.Direction)
	cashflow.BusinessType = AssetType(cf.BusinessType)
	p.parse("balance", cf.Balance, &cashflow.Balance)
	p.parse("business_time", cf.BusinessTime, &cashflow.BusinessTimestamp)
	return cashflow, p.errs.ToError()
}

type cashflowResp struct {
	statusResp
	Data struct {
		List []*cashflow
	}
}

// GetCashflows get cash flows of an account.
func (c *TradeClient) GetCashflows(r *CashflowReq) ([]*Cashflow, error) {
	if !r.IsAll {
		return c.getCashflows(r)
	}
	// Set default page and page size
	if r.Size == 0 {
		r.Size = 50
	}
	if r.Page == 0 {
		r.Page = 1
	}
	var cashflows []*Cashflow
	for {
		flows, err := c.getCashflows(r)
		glog.V(3).Infof("Get cashflows at page %d [Request: %+v, Error: %v]", r.Page, r, err)
		if err != nil {
			return nil, err
		}
		cashflows = append(cashflows, flows...)
		if len(flows) < int(r.Size) {
			break
		}
		r.Page++
	}
	return cashflows, nil
}

func (c *TradeClient) getCashflows(r *CashflowReq) ([]*Cashflow, error) {
	var resp cashflowResp
	if err := c.request(httpGET, cashflowURLPath, r.params(), nil, &resp); err != nil {
		return nil, err
	}
	return getCashflows(&resp)
}

func getCashflows(resp *cashflowResp) ([]*Cashflow, error) {
	if err := resp.CheckSuccess(); err != nil {
		return nil, err
	}
	var flows []*Cashflow
	var errs joinErrors
	for _, cf := range resp.Data.List {
		flow, err := cf.ToCashflow()
		if err != nil {
			errs.Add(fmt.Sprintf("balance %+v", cf), flow)
			continue
		}
		flows = append(flows, flow)
	}
	return flows, errs.ToError()
}

type FundPosition struct {
	Symbol                 Symbol
	SymbolName             string
	Currency               string
	HoldingUnits           float64
	CurrentNetAssetValue   float64
	CostNetAssetValue      float64
	NetAssetValueTimestamp int64
	AccountChannel         string
}

type fundPosition struct {
	Symbol               Symbol
	SymbolName           string `json:"symbol_name"`
	Currency             string
	HoldingUnits         string `json:"holding_units"`
	CurrentNetAssetValue string `json:"current_net_asset_value"`
	CostNetAssetValue    string `json:"cost_net_asset_value"`
	NetAssetValueDay     string `json:"net_asset_value_day"`
}

func (fp *fundPosition) toFundPosition(accountChannel string) (*FundPosition, error) {
	pos := &FundPosition{
		Symbol: fp.Symbol, SymbolName: fp.SymbolName, Currency: fp.Currency, AccountChannel: accountChannel,
	}
	p := &parser{}
	pos.HoldingUnits = p.parseFloat("holding_units", fp.HoldingUnits)
	pos.CurrentNetAssetValue = p.parseFloat("current_net_asset_value", fp.CurrentNetAssetValue)
	pos.CostNetAssetValue = p.parseFloat("cost_net_asset_value", fp.CostNetAssetValue)
	pos.NetAssetValueTimestamp = p.parseInt("net_asset_value_day", fp.NetAssetValueDay)
	return pos, p.errs.ToError()
}

type fundPositionResp struct {
	statusResp
	Data struct {
		List []struct {
			AccountChannel string         `json:"account_channel"`
			Funds          []fundPosition `json:"fund_info"`
		}
	}
}

// GetFundPositions gets fund positions. Fund symbols are optional.
func (c *TradeClient) GetFundPositions(fundSymbols ...Symbol) ([]*FundPosition, error) {
	vals := url.Values{}
	for _, s := range fundSymbols {
		vals.Add("symbol", string(s))
	}
	var resp fundPositionResp
	if err := c.request(httpGET, fundPositionURLPath, vals, nil, &resp); err != nil {
		return nil, err
	}
	return getFundPositions(&resp)
}

func getFundPositions(resp *fundPositionResp) ([]*FundPosition, error) {
	if err := resp.CheckSuccess(); err != nil {
		return nil, err
	}
	var positions []*FundPosition
	var errs joinErrors
	for _, funds := range resp.Data.List {
		for _, fp := range funds.Funds {
			pos, err := fp.toFundPosition(funds.AccountChannel)
			if err != nil {
				errs.Add(fmt.Sprintf("fund position %+v", fp), err)
				continue
			}
			positions = append(positions, pos)
		}
	}
	return positions, errs.ToError()
}

type StockPosition struct {
	Symbol            Symbol
	SymbolName        string
	Currency          string
	Quantity          uint64
	AvailableQuantity int64
	// CostPrice is cost price according to the TradeClient's choice of average purchase or diluted cost
	CostPrice      float64
	AccountChannel string
}

type stockPosition struct {
	Symbol     Symbol
	SymbolName string `json:"symbol_name"`
	Currency   string
	// Note: The schema for stock position wrongly use 'quality' to mean 'quantity', same for 'available_quality'.
	Quantity          string `json:"quality"`
	AvailableQuantity string `json:"available_quality"`
	CostPrice         string `json:"cost_price"`
}

func (sp *stockPosition) toPosition(accountChannel string) (*StockPosition, error) {
	pos := &StockPosition{
		Symbol: sp.Symbol, SymbolName: sp.SymbolName, Currency: sp.Currency, AccountChannel: accountChannel,
	}
	p := parser{}
	pos.Quantity = uint64(p.parseInt("quantity", sp.Quantity))
	availQty := p.parseFloat("available_quality", sp.AvailableQuantity)
	pos.AvailableQuantity = int64(availQty)
	if sp.CostPrice != "" {
		pos.CostPrice = p.parseFloat("cost_price", sp.CostPrice)
	}
	return pos, p.errs.ToError()
}

type stockPositionResp struct {
	statusResp
	Data struct {
		List []struct {
			AccountChannel string
			Stocks         []stockPosition `json:"stock_info"`
		}
	}
}

// GetStockPositions gets stock positions. If stock symbols are provided, only the positions matching the symbols will be returned.
// If stocks symbols are left empty, all account positions will be returned.
func (c *TradeClient) GetStockPositions(symbols ...Symbol) ([]*StockPosition, error) {
	vals := url.Values{}
	for _, s := range symbols {
		vals.Add("symbol", string(s))
	}
	var resp stockPositionResp
	if err := c.request(httpGET, stockPositionURLPath, vals, nil, &resp); err != nil {
		return nil, err
	}
	return getStockPositions(&resp)
}

func getStockPositions(resp *stockPositionResp) ([]*StockPosition, error) {
	if err := resp.CheckSuccess(); err != nil {
		return nil, err
	}
	var positions []*StockPosition
	var errs joinErrors
	for _, stocks := range resp.Data.List {
		for _, sp := range stocks.Stocks {
			pos, err := sp.toPosition(stocks.AccountChannel)
			if err != nil {
				errs.Add(fmt.Sprintf("stock position %+v", sp), err)
				continue
			}
			positions = append(positions, pos)
		}
	}
	return positions, errs.ToError()
}

type MarginRatio struct {
	InitRatio        float64
	MaintenanceRatio float64
	ForcedSaleRatio  float64
}

func (c *TradeClient) GetMarginRatio(symbol Symbol) (*MarginRatio, error) {
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Im string `json:"im_factor"`
			Mm string `json:"mm_factor"`
			Fm string `json:"fm_factor"`
		}
	}
	vals := url.Values{}
	vals.Add("symbol", string(symbol))
	if err := c.request(httpGET, marginRatioURLPath, vals, nil, &resp); err != nil {
		return nil, fmt.Errorf("error getting margin ratio for security %s: %v", symbol, err)
	}
	p := &parser{}
	mr := &MarginRatio{
		InitRatio:        p.parseFloat("im_factor", resp.Data.Im),
		MaintenanceRatio: p.parseFloat("mm_factor", resp.Data.Mm),
		ForcedSaleRatio:  p.parseFloat("fm_factor", resp.Data.Fm),
	}
	if err := p.Error(); err != nil {
		return nil, fmt.Errorf("Error parsing returned data for getting margin ratio for security %s: %v", symbol, err)
	}
	return mr, nil
}
