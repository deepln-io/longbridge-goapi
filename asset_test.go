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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsset(t *testing.T) {
	assert := assert.New(t)
	var resp balancesResp
	assert.Nil(decodeResp(strings.NewReader(`{
  "code": 0,
  "data": {
    "list": [
      {
        "total_cash": "1759070010.72",
        "max_finance_amount": "977582000",
        "remaining_finance_amount": "0",
        "risk_level": "1",
        "margin_call": "2598051051.50",
        "currency": "HKD",
        "cash_infos": [
          {
            "withdraw_cash": "97592.30",
            "available_cash": "195902464.37",
            "frozen_cash": "11579339.13",
            "settling_cash": "207288537.81",
            "currency": "HKD"
          }
        ]
      }
    ]
  }
}`), &resp))
	bals, err := getBalances(&resp)
	assert.Nil(err)
	assert.EqualValues(&Balance{
		TotalCash: 1759070010.72, MaxFinanceAmount: 977582000, RemainFinanceAmount: 0, RiskLevel: 1, MarginCall: 2598051051.5, Currency: "HKD",
		Cashes: []*Cash{
			{Withdraw: 97592.30, Available: 195902464.37, Frozen: 11579339.13, Settling: 207288537.81, Currency: "HKD"},
		},
	}, bals[0])

	assert.Equal("end_time=100&start_time=1", (&CashflowReq{StartTimestamp: 1, EndTimestamp: 100}).params().Encode())
	assert.Equal("business_type=2&end_time=100&page=2&size=200&start_time=1&symbol=A.US", (&CashflowReq{StartTimestamp: 1, EndTimestamp: 100, BusinessType: 2, Symbol: "A.US", Page: 2, Size: 200}).params().Encode())

	var cfResp cashflowResp
	assert.Nil(decodeResp(strings.NewReader(`{
  "code": 0,
  "data": {
    "list": [
      {
        "transaction_flow_name": "BuyContract-Stocks",
        "direction": 1,
		"business_type": 1,
        "balance": "-248.60",
        "currency": "USD",
        "business_time": "1621507957",
        "symbol": "AAPL.US",
        "description": "AAPL"
      }
    ]
  }
}`), &cfResp))
	cashflows, err := getCashflows(&cfResp)
	assert.Nil(err)
	assert.Equal(&Cashflow{TransactionFlowName: "BuyContract-Stocks", Direction: Outflow, BusinessType: CashAsset, Balance: -248.6, Currency: "USD", BusinessTimestamp: 1621507957, Symbol: "AAPL.US", Description: "AAPL"}, cashflows[0])

	var fpResp fundPositionResp
	assert.Nil(decodeResp(strings.NewReader(`{
  "code": 0,
  "data": {
    "list": [
      {
        "account_channel": "lb",
        "fund_info": [
          {
            "symbol": "HK0000447943",
            "symbol_name": "GE",
            "currency": "USD",
            "holding_units": "5.000",
            "current_net_asset_value": "1.020",
            "cost_net_asset_value": "0.10",
            "net_asset_value_day": "1600"
          }
        ]
      }
    ]
  }
}`), &fpResp))
	fpositions, err := getFundPositions(&fpResp)
	assert.Nil(err)
	assert.Equal(&FundPosition{Symbol: "HK0000447943", SymbolName: "GE", Currency: "USD", HoldingUnits: 5, CurrentNetAssetValue: 1.02, CostNetAssetValue: 0.1, NetAssetValueTimestamp: 1600, AccountChannel: "lb"}, fpositions[0])

	var spResp stockPositionResp
	assert.Nil(decodeResp(strings.NewReader(`{
  "code": 0,
  "data": {
    "list": [
      {
        "account_channel": "lb",
        "stock_info": [
          {
            "symbol": "700.HK",
            "symbol_name": "TENCENT",
            "currency": "HK",
            "quantity": "650",
            "available_quantity": "-450",
            "cost_price": "457.53"
          },
          {
            "symbol": "NOK.US",
            "symbol_name": "Nokia",
            "currency": "US",
            "quantity": "1"
          }
        ]
      }
    ]
  }
}`), &spResp))
	spositions, err := getStockPositions(&spResp)
	assert.Nil(err)
	assert.Equal([]*StockPosition{
		{Symbol: "700.HK", SymbolName: "TENCENT", Currency: "HK", Quantity: 650, AvailableQuantity: -450, CostPrice: 457.53},
		{Symbol: "NOK.US", SymbolName: "Nokia", Currency: "US", Quantity: 1},
	}, spositions)
}
