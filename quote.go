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
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/deepln-io/longbridge-goapi/internal/pb/quote"
	"github.com/deepln-io/longbridge-goapi/internal/protocol"

	"github.com/golang/glog"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

const (
	defaultQuoteAPITimeout = 10 * time.Second
)

type TradeStatus = quote.TradeStatus

const (
	TradeStatusNormal             = TradeStatus(quote.TradeStatus_NORMAL)
	TradeStatusHalted             = TradeStatus(quote.TradeStatus_HALTED)
	TradeStatusDelisted           = TradeStatus(quote.TradeStatus_DELISTED)
	TradeStatusFuse               = TradeStatus(quote.TradeStatus_FUSE)
	TradeStatusPrepareList        = TradeStatus(quote.TradeStatus_PREPARE_LIST)
	TradeStatusCodeMoved          = TradeStatus(quote.TradeStatus_CODE_MOVED)
	TradeStatusToBeOpened         = TradeStatus(quote.TradeStatus_TO_BE_OPENED)
	TradeStatusSlitStockHalts     = TradeStatus(quote.TradeStatus_SPLIT_STOCK_HALTS)
	TradeStatusExpired            = TradeStatus(quote.TradeStatus_EXPIRED)
	TradeStatusWarrantPrepareList = TradeStatus(quote.TradeStatus_WARRANT_PREPARE_LIST)
	TradeStatusSuspendTrade       = TradeStatus(quote.TradeStatus_SUSPEND_TRADE)
)

type Language int32

const (
	SimplifiedChinese  Language = 0
	English            Language = 1
	TraditionalChinese Language = 2
)

const watchedGroupURLPath urlPath = "/v1/watchlist/groups"

type FinanceIndex = quote.CalcIndex

const (
	IndexUnknown                 FinanceIndex = quote.CalcIndex_CALCINDEX_UNKNOWN
	IndexLastDone                FinanceIndex = quote.CalcIndex_CALCINDEX_LAST_DONE
	IndexChangeVal               FinanceIndex = quote.CalcIndex_CALCINDEX_CHANGE_VAL
	IndexChangeRate              FinanceIndex = quote.CalcIndex_CALCINDEX_CHANGE_RATE
	IndexVolume                  FinanceIndex = quote.CalcIndex_CALCINDEX_VOLUME
	IndexTurnover                FinanceIndex = quote.CalcIndex_CALCINDEX_TURNOVER
	IndexYtdChange_RATE          FinanceIndex = quote.CalcIndex_CALCINDEX_YTD_CHANGE_RATE
	IndexTurnoverRate            FinanceIndex = quote.CalcIndex_CALCINDEX_TURNOVER_RATE
	IndexTotalMarket_VALUE       FinanceIndex = quote.CalcIndex_CALCINDEX_TOTAL_MARKET_VALUE
	IndexCapitalFlow             FinanceIndex = quote.CalcIndex_CALCINDEX_CAPITAL_FLOW
	IndexAmplitude               FinanceIndex = quote.CalcIndex_CALCINDEX_AMPLITUDE
	IndexVolumeRatio             FinanceIndex = quote.CalcIndex_CALCINDEX_VOLUME_RATIO
	IndexPeTtm_RATIO             FinanceIndex = quote.CalcIndex_CALCINDEX_PE_TTM_RATIO
	IndexPbRatio                 FinanceIndex = quote.CalcIndex_CALCINDEX_PB_RATIO
	IndexDividendRatio_TTM       FinanceIndex = quote.CalcIndex_CALCINDEX_DIVIDEND_RATIO_TTM
	IndexFiveDay_CHANGE_RATE     FinanceIndex = quote.CalcIndex_CALCINDEX_FIVE_DAY_CHANGE_RATE
	IndexTenDay_CHANGE_RATE      FinanceIndex = quote.CalcIndex_CALCINDEX_TEN_DAY_CHANGE_RATE
	IndexHalfYear_CHANGE_RATE    FinanceIndex = quote.CalcIndex_CALCINDEX_HALF_YEAR_CHANGE_RATE
	IndexFiveMinutes_CHANGE_RATE FinanceIndex = quote.CalcIndex_CALCINDEX_FIVE_MINUTES_CHANGE_RATE
	IndexExpiryDate              FinanceIndex = quote.CalcIndex_CALCINDEX_EXPIRY_DATE
	IndexStrikePrice             FinanceIndex = quote.CalcIndex_CALCINDEX_STRIKE_PRICE
	IndexUpperStrike_PRICE       FinanceIndex = quote.CalcIndex_CALCINDEX_UPPER_STRIKE_PRICE
	IndexLowerStrike_PRICE       FinanceIndex = quote.CalcIndex_CALCINDEX_LOWER_STRIKE_PRICE
	IndexOutstandingQty          FinanceIndex = quote.CalcIndex_CALCINDEX_OUTSTANDING_QTY
	IndexOutstandingRatio        FinanceIndex = quote.CalcIndex_CALCINDEX_OUTSTANDING_RATIO
	IndexPremium                 FinanceIndex = quote.CalcIndex_CALCINDEX_PREMIUM
	IndexItmOtm                  FinanceIndex = quote.CalcIndex_CALCINDEX_ITM_OTM
	IndexImpliedVolatility       FinanceIndex = quote.CalcIndex_CALCINDEX_IMPLIED_VOLATILITY
	IndexWarrantDelta            FinanceIndex = quote.CalcIndex_CALCINDEX_WARRANT_DELTA
	IndexCallPrice               FinanceIndex = quote.CalcIndex_CALCINDEX_CALL_PRICE
	IndexToCall_PRICE            FinanceIndex = quote.CalcIndex_CALCINDEX_TO_CALL_PRICE
	IndexEffectiveLeverage       FinanceIndex = quote.CalcIndex_CALCINDEX_EFFECTIVE_LEVERAGE
	IndexLeverageRatio           FinanceIndex = quote.CalcIndex_CALCINDEX_LEVERAGE_RATIO
	IndexConversionRatio         FinanceIndex = quote.CalcIndex_CALCINDEX_CONVERSION_RATIO
	IndexBalancePoint            FinanceIndex = quote.CalcIndex_CALCINDEX_BALANCE_POINT
	IndexOpenInterest            FinanceIndex = quote.CalcIndex_CALCINDEX_OPEN_INTEREST
	IndexDelta                   FinanceIndex = quote.CalcIndex_CALCINDEX_DELTA
	IndexGamma                   FinanceIndex = quote.CalcIndex_CALCINDEX_GAMMA
	IndexTheta                   FinanceIndex = quote.CalcIndex_CALCINDEX_THETA
	IndexVega                    FinanceIndex = quote.CalcIndex_CALCINDEX_VEGA
	IndexRho                     FinanceIndex = quote.CalcIndex_CALCINDEX_RHO
)

type OrderBook struct {
	Position int32
	Price    float64
	Volume   int64
	OrderNum int64
}

type OrderBookList struct {
	Symbol Symbol
	Bid    []*OrderBook
	Ask    []*OrderBook
}

type Broker struct {
	Position int32
	IDs      []int32
}

type BrokerQueue struct {
	Symbol Symbol
	Bid    []*Broker
	Ask    []*Broker
}

type TradeDir int32

const (
	DirNeutral = TradeDir(0)
	DirDown    = TradeDir(1)
	DirUp      = TradeDir(2)
)

type Ticker struct {
	Price        float64
	Volume       int64
	Timestamp    int64
	TradeType    string
	Dir          TradeDir
	TradeSession TradeSessionType
}

type TradeSessionType = quote.TradeSession

const (
	NormalTradingSession = TradeSessionType(quote.TradeSession_NORMAL_TRADE)
	PreTradingSession    = TradeSessionType(quote.TradeSession_PRE_TRADE)
	PostTradingSession   = TradeSessionType(quote.TradeSession_POST_TRADE)
)

type PushQuote struct {
	Symbol          Symbol
	Sequence        int64
	LastDone        float64
	Open            float64
	High            float64
	Low             float64
	Timestamp       int64
	Volume          int64
	Turnover        float64
	TradeStatus     TradeStatus
	TradeSession    TradeSessionType
	CurrentVolume   int64
	CurrentTurnover float64
}

type PushOrderBook struct {
	Symbol   Symbol
	Sequence int64
	Bid      []*OrderBook
	Ask      []*OrderBook
}

type PushBrokers struct {
	Symbol   Symbol
	Sequence int64
	Bid      []*Broker
	Ask      []*Broker
}

type PushTickers struct {
	Symbol   Symbol
	Sequence int64
	Tickers  []*Ticker
}

type PreMarketQuote struct {
	LastDone  float64
	Timestamp int64
	Volume    int64
	Turnover  float64
	High      float64
	Low       float64
	PrevClose float64
}

type PostMarketQuote struct {
	LastDone  float64
	Timestamp int64
	Volume    int64
	Turnover  float64
	High      float64
	Low       float64
	PrevClose float64
}

type StaticInfo struct {
	Symbol            Symbol
	NameCn            string
	NameEn            string
	NameHk            string
	Exchange          string
	Currency          string
	LotSize           int32
	TotalShares       int64
	CirculatingShares int64
	HkShares          int64
	Eps               float64
	EpsTtm            float64
	Bps               float64
	DividendYield     float64
	StockDerivatives  []int32
	Board             string
}

type IntradayLine struct {
	Price     float64
	Timestamp int64
	Volume    int64
	Turnover  float64
	AvgPrice  float64
}

type RealTimeQuote struct {
	Symbol          Symbol
	LastDone        float64
	PrevClose       float64
	Open            float64
	High            float64
	Low             float64
	Timestamp       int64
	Volume          int64
	Turnover        float64
	PreMarketQuote  *PreMarketQuote
	PostMarketQuote *PostMarketQuote
}

type KLineType int32

const (
	KLine1M    = KLineType(quote.Period_ONE_MINUTE)
	KLine5M    = KLineType(quote.Period_FIVE_MINUTE)
	KLine15M   = KLineType(quote.Period_FIFTEEN_MINUTE)
	KLine30M   = KLineType(quote.Period_THIRTY_MINUTE)
	KLine60M   = KLineType(quote.Period_SIXTY_MINUTE)
	KLineDay   = KLineType(quote.Period_DAY)
	KLineWeek  = KLineType(quote.Period_WEEK)
	KLineMonth = KLineType(quote.Period_MONTH)
	KLineYear  = KLineType(quote.Period_YEAR)
)

type KLine struct {
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
	Turnover  float64
	Timestamp int64
}

type AdjustType int32

const (
	AdjustNone = AdjustType(quote.AdjustType_NO_ADJUST)
	AdjustFwd  = AdjustType(quote.AdjustType_FORWARD_ADJUST)
)

type OptionExtend struct {
	ImpliedVolatility    float64
	OpenInterest         int64
	ExpiryDate           string // YYMMDD
	StrikePrice          float64
	ContractMultiplier   float64
	ContractType         string
	ContractSize         float64
	Direction            string
	HistoricalVolatility float64
	UnderlyingSymbol     Symbol
}

type StrikePriceInfo struct {
	Price      float64
	CallSymbol string
	PutSymbol  string
	Standard   bool
}

type Issuer struct {
	ID     int32
	NameCn string
	NameEn string
	NameHk string
}

type RealtimeOptionQuote struct {
	Symbol       Symbol
	LastDone     float64
	PrevClose    float64
	Open         float64
	High         float64
	Low          float64
	Timestamp    int64
	Volume       int64
	Turnover     float64
	OptionExtend *OptionExtend
}

type WarrantExtended struct {
	ImpliedVolatility float64
	ExpiryDate        string
	LastTradeDate     string
	OutstandingRatio  float64
	OutstandingQty    int64
	ConversionRatio   float64
	Category          string
	StrikePrice       float64
	UpperStrikePrice  float64
	LowerStrikePrice  float64
	CallPrice         float64
	UnderlyingSymbol  string
}

type RealtimeWarrantQuote struct {
	Symbol        Symbol
	LastDone      float64
	PrevClose     float64
	Open          float64
	High          float64
	Low           float64
	Timestamp     int64
	Volume        int64
	Turnover      float64
	WarrantExtend *WarrantExtended
}

// WarrantFilter includes the search conditions for warrants. The field defintion can refer to
// https://open.longbridgeapp.com/en/docs/quote/pull/warrant-filter
type WarrantFilter struct {
	Symbol   Symbol
	Language Language

	SortBy     int32
	SortOrder  int32 // 0 Ascending 1 Desending
	SortOffset int32
	PageSize   int32 // Up to 500

	// The following are optional

	Type      []int32 // optional values: 0 - Call	1 - Put 2 - Bull 3 - Bear 4 - Inline
	IssuerIDs []int32

	// ExpiryDateType can have the following values.
	// 1 - Less than 3 months
	// 2 - 3 - 6 months
	// 3 - 6 - 12 months
	// 4 - greater than 12 months
	ExpiryDateType []int32

	// Optional values for PriceType
	// 1 - In bounds
	// 2 - Out bounds
	PriceType []int32

	// Optional values for Status:
	// 2 - Suspend trading
	// 3 - Papare List
	// 4 - Normal
	Status []int32
}

type Warrant struct {
	Symbol            Symbol
	Name              string
	LastDone          float64
	ChangeRate        float64
	ChangeVal         float64
	Turnover          float64
	ExpiryDate        string // YYYYMMDD
	StrikePrice       float64
	UpperStrikePrice  float64
	LowerStrikePrice  float64
	OutstandingQty    float64
	OutstandingRatio  float64
	Premium           float64
	ItmOtm            float64
	ImpliedVolatility float64
	Delta             float64
	CallPrice         float64
	EffectiveLeverage float64
	LeverageRatio     float64
	ConversionRatio   float64
	BalancePoint      float64
	State             string
}

type MarketSession struct {
	SessionType TradeSessionType
	BeginTime   int32 // The time is encoded with int as HHMM, e.g., 930 means 9:30am, and it is in the corresponding market timezone.
	EndTime     int32
}

type MarketTradePeriod struct {
	Market        Market
	TradeSessions []*MarketSession
}

type TradeDate struct {
	Date          string
	TradeDateType int32 // 0 full day, 1 morning only, 2 afternoon only(not happened before)
}

type IntradayCapFlow struct {
	Flow      float64
	Timestamp int64
}

type CapDistribution struct {
	Large  float64
	Medium float64
	Small  float64
}

type CapFlowDistribution struct {
	Symbol    Symbol
	Timestamp int64
	InFlow    CapDistribution
	OutFlow   CapDistribution
}

type SecurityFinanceIndex struct {
	Symbol                Symbol
	LastDone              float64
	ChangeVal             float64
	ChangeRate            float64
	Volume                int64
	Turnover              float64
	YtdChangeRate         float64
	TurnoverRate          float64
	TotalMarketValue      float64
	CapitalFlow           float64
	Amplitude             float64
	VolumeRatio           float64
	PeTtmRatio            float64
	PbRatio               float64
	DividendRatioTtm      float64
	FiveDayChangeRate     float64
	TenDayChangeRate      float64
	HalfYearChangeRate    float64
	FiveMinutesChangeRate float64
	ExpiryDate            string
	StrikePrice           float64
	UpperStrikePrice      float64
	LowerStrikePrice      float64
	OutstandingQty        int64
	OutstandingRatio      float64
	Premium               float64
	ItmOtm                float64
	ImpliedVolatility     float64
	WarrantDelta          float64
	CallPrice             float64
	ToCallPrice           float64
	EffectiveLeverage     float64
	LeverageRatio         float64
	ConversionRatio       float64
	BalancePoint          float64
	OpenInterest          int64
	Delta                 float64
	Gamma                 float64
	Theta                 float64
	Vega                  float64
	Rho                   float64
}

type SubscriptionType = quote.SubType

const (
	SubscriptionRealtimeQuote = SubscriptionType(quote.SubType_QUOTE)
	SubscriptionOrderBook     = SubscriptionType(quote.SubType_DEPTH)
	SubscriptionBrokerQueue   = SubscriptionType(quote.SubType_BROKERS)
	SubscriptionTicker        = SubscriptionType(quote.SubType_TRADE)
)

type QotSubscription struct {
	Symbol        Symbol
	Subscriptions []SubscriptionType
}

type WatchedSecurity struct {
	Symbol       Symbol
	Name         string
	WatchedPrice float64
	WatchedAt    int64
}

func (wg *WatchedSecurity) String() string {
	return fmt.Sprintf("Security %s (%s) %g on %v", wg.Symbol, wg.Name, wg.WatchedPrice, time.Unix(wg.WatchedAt, 0))
}

type WatchedGroup struct {
	ID         string
	Name       string
	Securities []*WatchedSecurity
}

// quoteLongConn is the connection for quote related APIs and push notification.
// To receive pushed quote data, set the callback OnXXX in the connection and call Enable(true) to start
// the connection to receive pushed quote data.
type quoteLongConn struct {
	*longConn
	// Map from subscription types to symbols, used for subscription restoration after reconnecting
	subsMu          sync.RWMutex
	subs            map[SubscriptionType][]Symbol
	OnPushQuote     func(q *PushQuote)
	OnPushOrderBook func(o *PushOrderBook)
	OnPushBrokers   func(b *PushBrokers)
	OnPushTickers   func(t *PushTickers)
}

func newQuoteLongConn(endPoint string, otpProvider otpProvider) *quoteLongConn {
	c := &quoteLongConn{longConn: newLongConn(endPoint, otpProvider), subs: make(map[quote.SubType][]Symbol)}
	c.onPush = c.handlePushPkg
	c.longConn.recover = c.restoreSubscriptions
	// long bridge limits: max 10 requests/second, and max 5 concurrent calls.
	c.longConn.limiters = []*rate.Limiter{
		rate.NewLimiter(10, 5), // 10 tokens/second, bucket size 5
	}
	for _, limiter := range c.longConn.limiters {
		limiter.WaitN(context.Background(), limiter.Burst()) // Clear the initial buckets since initial bucket is full
	}
	return c
}

func (c *quoteLongConn) updateSubscriptions(symbols []Symbol, subTypes []SubscriptionType, add bool) {
	c.subsMu.Lock()
	defer c.subsMu.Unlock()
	for _, st := range subTypes {
		existings := c.subs[st]
		m := make(map[Symbol]bool)
		for _, sym := range existings {
			m[sym] = true
		}
		for _, sym := range symbols {
			if add {
				m[sym] = true
			} else if m[sym] {
				delete(m, sym)
			}
		}
		var ns []Symbol
		for sym := range m {
			ns = append(ns, sym)
		}
		c.subs[st] = ns
	}
}

func (c *quoteLongConn) restoreSubscriptions() {
	c.subsMu.RLock()
	m := make(map[SubscriptionType][]Symbol, len(c.subs))
	for st, symbols := range c.subs {
		m[st] = symbols
	}
	c.subsMu.RUnlock()
	go func() {
		for st, symbols := range m {
			c.SubscribePush(symbols, []quote.SubType{st}, true)
		}
	}()
}

func (c *quoteLongConn) GetStaticInfo(symbols []Symbol) ([]*StaticInfo, error) {
	header := c.getReqHeader(protocol.CmdInfo)
	var resp quote.SecurityStaticInfoResponse
	if err := c.Call("security_static_info", header,
		&quote.MultiSecurityRequest{Symbol: makeStringSlice(symbols)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ss []*StaticInfo
	for _, s := range resp.SecuStaticInfo {
		p := &parser{}
		ss = append(ss, &StaticInfo{
			Symbol:            Symbol(s.Symbol),
			NameCn:            s.NameCn,
			NameEn:            s.NameEn,
			NameHk:            s.NameHk,
			Exchange:          s.Exchange,
			Currency:          s.Currency,
			LotSize:           s.LotSize,
			TotalShares:       s.CirculatingShares,
			CirculatingShares: s.CirculatingShares,
			HkShares:          s.HkShares,
			Eps:               p.parseFloat("eps", s.Eps),
			EpsTtm:            p.parseFloat("eps_ttm", s.EpsTtm),
			Bps:               p.parseFloat("bps", s.Bps),
			DividendYield:     p.parseFloat("dividend_yield", s.DividendYield),
			StockDerivatives:  s.StockDerivatives,
			Board:             s.Board,
		})
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error in static info response for %q format: %v", s.Symbol, err)
		}
	}
	return ss, nil
}

func (c *quoteLongConn) GetRealtimeQuote(symbols []Symbol) ([]*RealTimeQuote, error) {
	header := c.getReqHeader(protocol.CmdRealtimeQuote)
	var resp quote.SecurityQuoteResponse
	if err := c.Call("security_quote", header,
		&quote.MultiSecurityRequest{Symbol: makeStringSlice(symbols)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var qs []*RealTimeQuote
	for _, q := range resp.SecuQuote {
		log.Printf("q: %#v\n", q)
		p := &parser{}
		rq := &RealTimeQuote{
			Symbol:    Symbol(q.Symbol),
			LastDone:  p.parseFloat("last_done", q.LastDone),
			PrevClose: p.parseFloat("prev_close", q.PrevClose),
			Open:      p.parseFloat("open", q.Open),
			High:      p.parseFloat("high", q.High),
			Low:       p.parseFloat("low", q.Low),
			Timestamp: q.Timestamp,
			Volume:    q.Volume,
			Turnover:  p.parseFloat("turnover", q.Turnover),
		}
		if q.PreMarketQuote != nil {
			rq.PreMarketQuote = &PreMarketQuote{
				LastDone:  p.parseFloat("last_done", q.PreMarketQuote.LastDone),
				Timestamp: q.PreMarketQuote.Timestamp,
				Volume:    q.PreMarketQuote.Volume,
				Turnover:  p.parseFloat("turnover", q.PreMarketQuote.Turnover),
				High:      p.parseFloat("high", q.PreMarketQuote.High),
				Low:       p.parseFloat("low", q.PreMarketQuote.Low),
				PrevClose: p.parseFloat("prev_close", q.PreMarketQuote.PrevClose),
			}
		}
		if q.PostMarketQuote != nil {
			rq.PostMarketQuote = &PostMarketQuote{
				LastDone:  p.parseFloat("last_done", q.PostMarketQuote.LastDone),
				Timestamp: q.PostMarketQuote.Timestamp,
				Volume:    q.PostMarketQuote.Volume,
				Turnover:  p.parseFloat("turnover", q.PostMarketQuote.Turnover),
				High:      p.parseFloat("high", q.PostMarketQuote.High),
				Low:       p.parseFloat("low", q.PostMarketQuote.Low),
				PrevClose: p.parseFloat("prev_close", q.PostMarketQuote.PrevClose),
			}
		}
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error in real time quote for %q: %v", q.Symbol, err)
		}
		qs = append(qs, rq)
	}
	return qs, nil
}

func (c *quoteLongConn) GetRealtimeOptionQuote(symbols []Symbol) ([]*RealtimeOptionQuote, error) {
	header := c.getReqHeader(protocol.CmdRealtimeOptionQuote)
	var resp quote.OptionQuoteResponse
	if err := c.Call("option_quote", header,
		&quote.MultiSecurityRequest{Symbol: makeStringSlice(symbols)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var qs []*RealtimeOptionQuote
	for _, q := range resp.SecuQuote {
		p := &parser{}
		ro := &RealtimeOptionQuote{
			Symbol:    Symbol(q.Symbol),
			LastDone:  p.parseFloat("last_done", q.LastDone),
			PrevClose: p.parseFloat("prev_close", q.PrevClose),
			Open:      p.parseFloat("open", q.Open),
			High:      p.parseFloat("high", q.High),
			Low:       p.parseFloat("low", q.Low),
			Timestamp: q.Timestamp,
			Volume:    q.Volume,
			Turnover:  p.parseFloat("turnover", q.Turnover),
		}
		if q.OptionExtend != nil {
			ro.OptionExtend = &OptionExtend{
				ImpliedVolatility:    p.parseFloat("implied_volatility", q.OptionExtend.ImpliedVolatility),
				OpenInterest:         q.OptionExtend.OpenInterest,
				ExpiryDate:           q.OptionExtend.ExpiryDate,
				StrikePrice:          p.parseFloat("strike_price", q.OptionExtend.StrikePrice),
				ContractMultiplier:   p.parseFloat("contract_multiplier", q.OptionExtend.ContractMultiplier),
				ContractType:         q.OptionExtend.ContractType,
				ContractSize:         p.parseFloat("contract_size", q.OptionExtend.ContractSize),
				Direction:            q.OptionExtend.Direction,
				HistoricalVolatility: p.parseFloat("historical_volatility", q.OptionExtend.HistoricalVolatility),
				UnderlyingSymbol:     Symbol(q.OptionExtend.UnderlyingSymbol),
			}
		}
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error in real time option: %v", err)
		}
		qs = append(qs, ro)
	}
	return qs, nil
}

func (c *quoteLongConn) GetRealtimeWarrantQuote(symbols []Symbol) ([]*RealtimeWarrantQuote, error) {
	header := c.getReqHeader(protocol.CmdRealtimeWarrantQuote)
	var resp quote.WarrantQuoteResponse
	if err := c.Call("warrant_quote", header,
		&quote.MultiSecurityRequest{Symbol: makeStringSlice(symbols)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ws []*RealtimeWarrantQuote
	for _, q := range resp.SecuQuote {
		p := &parser{}
		w := &RealtimeWarrantQuote{
			Symbol:    Symbol(q.Symbol),
			LastDone:  p.parseFloat("last_done", q.LastDone),
			PrevClose: p.parseFloat("prev_close", q.PrevClose),
			Open:      p.parseFloat("open", q.Open),
			High:      p.parseFloat("high", q.High),
			Low:       p.parseFloat("low", q.Low),
			Timestamp: q.Timestamp,
			Volume:    q.Volume,
			Turnover:  p.parseFloat("turnover", q.Turnover),
		}
		if q.WarrantExtend != nil {
			w.WarrantExtend = &WarrantExtended{
				ImpliedVolatility: p.parseFloat("implied_volatility", q.WarrantExtend.ImpliedVolatility),
				ExpiryDate:        q.WarrantExtend.ExpiryDate,
				LastTradeDate:     q.WarrantExtend.LastTradeDate,
				OutstandingRatio:  p.parseFloat("outstanding_rate", q.WarrantExtend.OutstandingRatio),
				OutstandingQty:    q.WarrantExtend.OutstandingQty,
				ConversionRatio:   p.parseFloat("conversion_ratio", q.WarrantExtend.ConversionRatio),
				Category:          q.WarrantExtend.Category,
				StrikePrice:       p.parseFloat("strike_price", q.WarrantExtend.StrikePrice),
				UpperStrikePrice:  p.parseFloat("upper_strike_price", q.WarrantExtend.UpperStrikePrice),
				LowerStrikePrice:  p.parseFloat("lower_strike_price", q.WarrantExtend.LowerStrikePrice),
				CallPrice:         p.parseFloat("call_price", q.WarrantExtend.CallPrice),
				UnderlyingSymbol:  q.WarrantExtend.UnderlyingSymbol,
			}
		}
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error in real time option: %v", err)
		}
		ws = append(ws, w)
	}
	return ws, nil
}

func (c *quoteLongConn) GetOrderBookList(symbol Symbol) (*OrderBookList, error) {
	header := c.getReqHeader(protocol.CmdOrderBook)
	var resp quote.SecurityDepthResponse
	if err := c.Call("security_depth", header,
		&quote.SecurityRequest{Symbol: string(symbol)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	ol := &OrderBookList{Symbol: Symbol(resp.Symbol)}
	p := &parser{}
	for _, ask := range resp.Ask {
		ol.Ask = append(ol.Ask, &OrderBook{
			Position: ask.Position,
			Price:    p.parseFloat("ask_price", ask.Price),
			Volume:   ask.Volume,
			OrderNum: ask.OrderNum,
		})
	}
	for _, bid := range resp.Bid {
		ol.Bid = append(ol.Ask, &OrderBook{
			Position: bid.Position,
			Price:    p.parseFloat("bid_price", bid.Price),
			Volume:   bid.Volume,
			OrderNum: bid.OrderNum,
		})
	}
	if err := p.Error(); err != nil {
		return nil, fmt.Errorf("error parsing security depth: %v", err)
	}
	return ol, nil
}

func (c *quoteLongConn) GetBrokerQueue(symbol Symbol) (*BrokerQueue, error) {
	header := c.getReqHeader(protocol.CmdBrokerQueue)
	var resp quote.SecurityBrokersResponse
	if err := c.Call("security_borkers", header,
		&quote.SecurityRequest{Symbol: string(symbol)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	b := &BrokerQueue{Symbol: Symbol(resp.Symbol)}
	for _, bid := range resp.BidBrokers {
		b.Bid = append(b.Bid, &Broker{Position: bid.Position, IDs: bid.BrokerIds})
	}
	for _, ask := range resp.AskBrokers {
		b.Ask = append(b.Ask, &Broker{Position: ask.Position, IDs: ask.BrokerIds})
	}
	return b, nil
}

type BrokerInfo struct {
	IDs         []int32
	NameEnglish string
	NameChinese string
	NameHK      string
}

func (c *quoteLongConn) GetBrokerInfo() ([]*BrokerInfo, error) {
	header := c.getReqHeader(protocol.CmdBrokerInfo)
	var resp quote.ParticipantBrokerIdsResponse
	if err := c.Call("participant_broker_ids", header, nil, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var bs []*BrokerInfo
	for _, b := range resp.ParticipantBrokerNumbers {
		bs = append(bs, &BrokerInfo{
			IDs:         b.BrokerIds,
			NameEnglish: b.ParticipantNameEn,
			NameChinese: b.ParticipantNameCn,
			NameHK:      b.ParticipantNameHk,
		})
	}
	return bs, nil
}

func (c *quoteLongConn) GetTickers(symbol Symbol, count int) ([]*Ticker, error) {
	header := c.getReqHeader(protocol.CmdTicker)
	var resp quote.SecurityTradeResponse
	if err := c.Call("security_trade", header,
		&quote.SecurityTradeRequest{Symbol: string(symbol), Count: int32(count)},
		&resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ts []*Ticker
	for _, t := range resp.Trades {
		p := &parser{}
		ts = append(ts, &Ticker{
			Price:        p.parseFloat("price", t.Price),
			Volume:       t.Volume,
			Timestamp:    t.Timestamp,
			TradeType:    t.TradeType,
			Dir:          TradeDir(t.Direction),
			TradeSession: TradeSessionType(t.TradeSession),
		})
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error parsing security trade for %q: %v", symbol, err)
		}
	}
	return ts, nil
}

func (c *quoteLongConn) GetIntradayLines(symbol Symbol) ([]*IntradayLine, error) {
	header := c.getReqHeader(protocol.CmdIntraday)
	var resp quote.SecurityIntradayResponse
	if err := c.Call("intraday", header, &quote.SecurityIntradayRequest{Symbol: string(symbol)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ls []*IntradayLine
	for _, l := range resp.Lines {
		p := &parser{}
		ls = append(ls, &IntradayLine{
			Price:     p.parseFloat("price", l.Price),
			Timestamp: l.Timestamp,
			Volume:    l.Volume,
			Turnover:  p.parseFloat("turnover", l.Turnover),
			AvgPrice:  p.parseFloat("avg_price", l.AvgPrice),
		})
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error parsing intraday line: %v", err)
		}
	}
	return ls, nil
}

func (c *quoteLongConn) GetKLines(symbol Symbol, klType KLineType, count int32, adj AdjustType) ([]*KLine, error) {
	header := c.getReqHeader(protocol.CmdKLine)
	var resp quote.SecurityCandlestickResponse
	if err := c.Call("candlestick", header, &quote.SecurityCandlestickRequest{
		Symbol:     string(symbol),
		Period:     quote.Period(klType),
		Count:      count,
		AdjustType: quote.AdjustType(adj),
	}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ks []*KLine
	for _, k := range resp.Candlesticks {
		p := &parser{}
		ks = append(ks, &KLine{
			Open:      p.parseFloat("open", k.Open),
			High:      p.parseFloat("high", k.High),
			Low:       p.parseFloat("low", k.Low),
			Close:     p.parseFloat("close", k.Close),
			Volume:    k.Volume,
			Turnover:  p.parseFloat("turnover", k.Turnover),
			Timestamp: k.Timestamp,
		})
		if err := p.Error(); err != nil {
			return nil, err
		}
	}
	return ks, nil
}

// GetOptionExpiryDates returns the expiry dates for a stock. The dates are encoded as YYMMDD, with timezone related to its corresponding market.
func (c *quoteLongConn) GetOptionExpiryDates(symbol Symbol) ([]string, error) {
	header := c.getReqHeader(protocol.CmdOptionExpiryDate)
	var resp quote.OptionChainDateListResponse
	if err := c.Call("option_chain_expiry_date_list", header,
		&quote.SecurityRequest{Symbol: string(symbol)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	return resp.ExpiryDate, nil
}

// GetOptionStrikePrices returns the strike price list for a stock's option chain on given expiry date (in YYYYMMDD format).
func (c *quoteLongConn) GetOptionStrikePrices(symbol Symbol, expiry string) ([]*StrikePriceInfo, error) {
	header := c.getReqHeader(protocol.CmdOptionDateStrikeInfo)
	var resp quote.OptionChainDateStrikeInfoResponse
	if err := c.Call("option_chain_dates_strike_price_info", header,
		&quote.OptionChainDateStrikeInfoRequest{Symbol: string(symbol), ExpiryDate: expiry},
		&resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ss []*StrikePriceInfo
	for _, s := range resp.StrikePriceInfo {
		p := &parser{}
		ss = append(ss, &StrikePriceInfo{
			Price:      p.parseFloat("price", s.Price),
			CallSymbol: s.CallSymbol,
			PutSymbol:  s.PutSymbol,
			Standard:   s.Standard,
		})
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error parsing strick price info: %v", err)
		}
	}
	return ss, nil
}

func (c *quoteLongConn) GetWarrantIssuers() ([]*Issuer, error) {
	header := c.getReqHeader(protocol.CmdWarrantIssuers)
	var resp quote.IssuerInfoResponse
	if err := c.Call("warrant_issuers", header, nil, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var is []*Issuer
	for _, i := range resp.IssuerInfo {
		is = append(is, &Issuer{
			ID:     i.Id,
			NameCn: i.NameCn,
			NameEn: i.NameEn,
			NameHk: i.NameHk,
		})
	}
	return is, nil
}

func (c *quoteLongConn) SearchWarrants(cond *WarrantFilter) ([]*Warrant, error) {
	header := c.getReqHeader(protocol.CmdSearchWarrant)
	var resp quote.WarrantFilterListResponse
	req := &quote.WarrantFilterListRequest{
		Symbol: string(cond.Symbol),
		FilterConfig: &quote.FilterConfig{
			SortBy:     cond.SortBy,
			SortOrder:  cond.SortOrder,
			SortOffset: cond.SortOffset,
			SortCount:  cond.PageSize,
			Type:       cond.Type,
			Issuer:     cond.IssuerIDs,
			ExpiryDate: cond.ExpiryDateType,
			PriceType:  cond.PriceType,
			Status:     cond.Status,
		},
		Language: int32(cond.Language),
	}
	if err := c.Call("warrant_filter", header, req, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ws []*Warrant
	for _, w := range resp.WarrantList {
		p := &parser{}
		ws = append(ws, &Warrant{
			Symbol:            Symbol(w.Symbol),
			Name:              w.Name,
			LastDone:          p.parseFloat("last_done", w.LastDone),
			ChangeRate:        p.parseFloat("change_rate", w.ChangeRate),
			ChangeVal:         p.parseFloat("change_val", w.ChangeVal),
			Turnover:          p.parseFloat("turnover", w.Turnover),
			ExpiryDate:        w.ExpiryDate,
			StrikePrice:       p.parseFloat("strike_price", w.StrikePrice),
			UpperStrikePrice:  p.parseFloat("upper_strike_price", w.UpperStrikePrice),
			LowerStrikePrice:  p.parseFloat("lower_strike_price", w.LowerStrikePrice),
			OutstandingQty:    p.parseFloat("outstanding_qty", w.OutstandingQty),
			OutstandingRatio:  p.parseFloat("outstanding_ratio", w.OutstandingRatio),
			Premium:           p.parseFloat("premium", w.Premium),
			ItmOtm:            p.parseFloat("item_otm", w.ItmOtm),
			ImpliedVolatility: p.parseFloat("implied_volatility", w.ImpliedVolatility),
			Delta:             p.parseFloat("delta", w.Delta),
			CallPrice:         p.parseFloat("call_price", w.CallPrice),
			EffectiveLeverage: p.parseFloat("effective_leverage", w.EffectiveLeverage),
			LeverageRatio:     p.parseFloat("leverage_ratio", w.LeverageRatio),
			ConversionRatio:   p.parseFloat("conversion_ratio", w.ConversionRatio),
			BalancePoint:      p.parseFloat("balance_point", w.BalancePoint),
			State:             w.State,
		})
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error parsing warant data for %q: %v", w.Symbol, err)
		}
	}
	return ws, nil
}

func (c *quoteLongConn) GetMarketTradePeriods() ([]*MarketTradePeriod, error) {
	header := c.getReqHeader(protocol.CmdMarketTradePeriod)
	var resp quote.MarketTradePeriodResponse
	if err := c.Call("market_trade_period", header, nil, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ms []*MarketTradePeriod
	for _, s := range resp.MarketTradeSession {
		m := &MarketTradePeriod{Market: Market(s.Market)}
		for _, ss := range s.TradeSession {
			m.TradeSessions = append(m.TradeSessions,
				&MarketSession{
					SessionType: TradeSessionType(ss.TradeSession),
					BeginTime:   ss.BegTime,
					EndTime:     ss.EndTime})
		}
		ms = append(ms, m)
	}
	return ms, nil
}

// GetTradeDates returns the trading days in given time range for the market. The begin and end are encoded as "20060102" format.
// The trading days are sorted in ascending time order.
// Note: The interval cannot be greater than one month.
// Only supports query data of the most recent year
func (c *quoteLongConn) GetTradeDates(market Market, begin string, end string) ([]TradeDate, error) {
	header := c.getReqHeader(protocol.CmdTradeDate)
	var resp quote.MarketTradeDayResponse
	if err := c.Call("market_trade_day", header,
		&quote.MarketTradeDayRequest{Market: string(market), BegDay: begin, EndDay: end},
		&resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ts []TradeDate
	for _, t := range resp.TradeDay {
		ts = append(ts, TradeDate{Date: t})
	}
	for _, t := range resp.HalfTradeDay {
		ts = append(ts, TradeDate{Date: t, TradeDateType: 1})
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i].Date < ts[j].Date })
	return ts, nil
}

func (c *quoteLongConn) GetSubscriptions() ([]*QotSubscription, error) {
	header := c.getReqHeader(protocol.CmdSubscription)
	var resp quote.SubscriptionResponse
	if err := c.Call("subscription", header, &quote.SubscribeRequest{}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var ss []*QotSubscription
	for _, s := range resp.SubList {
		qs := &QotSubscription{Symbol: Symbol(s.Symbol)}
		for _, st := range s.SubType {
			qs.Subscriptions = append(qs.Subscriptions, SubscriptionType(st))
		}
		ss = append(ss, qs)
	}
	return ss, nil
}

func (c *quoteLongConn) SubscribePush(symbols []Symbol, subTypes []SubscriptionType, needFirstPush bool) ([]*QotSubscription, error) {
	header := c.getReqHeader(protocol.CmdSubscribe)
	var resp quote.SubscriptionResponse
	if err := c.Call("subscribe", header, &quote.SubscribeRequest{
		Symbol:      makeStringSlice(symbols),
		SubType:     subTypes,
		IsFirstPush: needFirstPush,
	}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var qs []*QotSubscription
	for _, s := range resp.SubList {
		qs = append(qs, &QotSubscription{
			Symbol:        Symbol(s.Symbol),
			Subscriptions: s.SubType,
		})
	}
	c.updateSubscriptions(symbols, subTypes, true)
	return qs, nil
}

// Unsubscribe clears the subscription given by symbols. If symbols is empty and all is true,
// it will clear the requested subscriptions for all subscribed symbols.
func (c *quoteLongConn) UnsubscribePush(symbols []Symbol, subTypes []SubscriptionType, all bool) error {
	header := c.getReqHeader(protocol.CmdUnsubscribe)
	var resp quote.UnsubscribeResponse
	if err := c.Call("unsubscribe", header,
		&quote.UnsubscribeRequest{Symbol: makeStringSlice(symbols), SubType: subTypes, UnsubAll: all},
		&resp, defaultQuoteAPITimeout); err != nil {
		return err
	}
	c.updateSubscriptions(symbols, subTypes, false)
	return nil
}

func (c *quoteLongConn) GetIntradayCapFlows(symbol Symbol) ([]*IntradayCapFlow, error) {
	header := c.getReqHeader(protocol.CmdIntradayCapFlow)
	var resp quote.CapitalFlowIntradayResponse
	if err := c.Call("intraday_capflow", header,
		&quote.CapitalFlowIntradayRequest{Symbol: string(symbol)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var flows []*IntradayCapFlow
	for _, line := range resp.CapitalFlowLines {
		p := &parser{}
		flows = append(flows, &IntradayCapFlow{
			Flow:      p.parseFloat("flow", line.Inflow),
			Timestamp: line.Timestamp,
		})
		if err := p.Error(); err != nil {
			return nil, fmt.Errorf("error parsing intraday capital flow : %v", err)
		}
	}
	return flows, nil
}

func (c *quoteLongConn) GetIntradayCapFlowDistribution(symbol Symbol) (*CapFlowDistribution, error) {
	header := c.getReqHeader(protocol.CmdIntradayCapFlowDistribution)
	var resp quote.CapitalDistributionResponse
	if err := c.Call("intraday_capflow_distribution", header,
		&quote.SecurityRequest{Symbol: string(symbol)}, &resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	p := &parser{}
	cd := &CapFlowDistribution{
		Symbol:    symbol,
		Timestamp: resp.Timestamp,
		InFlow: CapDistribution{
			Large:  p.parseFloat("inflow_large", resp.CapitalIn.Large),
			Medium: p.parseFloat("inflow_medium", resp.CapitalIn.Medium),
			Small:  p.parseFloat("inflow_small", resp.CapitalIn.Small),
		},
		OutFlow: CapDistribution{
			Large:  p.parseFloat("outflow_large", resp.CapitalOut.Large),
			Medium: p.parseFloat("outflow_medium", resp.CapitalOut.Medium),
			Small:  p.parseFloat("outflow_small", resp.CapitalOut.Small),
		},
	}
	if err := p.Error(); err != nil {
		return nil, err
	}
	return cd, nil
}

func (c *quoteLongConn) GetFinanceIndices(symbols []Symbol, indices []FinanceIndex) ([]*SecurityFinanceIndex, error) {
	header := c.getReqHeader(protocol.CmdSecurityFinanceIndex)
	var resp quote.SecurityCalcQuoteResponse
	if err := c.Call("finance_indices", header,
		&quote.SecurityCalcQuoteRequest{Symbols: makeStringSlice(symbols), CalcIndex: indices},
		&resp, defaultQuoteAPITimeout); err != nil {
		return nil, err
	}
	var fs []*SecurityFinanceIndex
	for _, s := range resp.SecurityCalcIndex {
		p := &parser{}
		f := &SecurityFinanceIndex{
			Symbol:                Symbol(s.Symbol),
			LastDone:              p.parseFloat("last_done", s.LastDone),
			ChangeVal:             p.parseFloat("change_val", s.ChangeVal),
			ChangeRate:            p.parseFloat("change_rate", s.ChangeRate),
			Volume:                s.Volume,
			Turnover:              p.parseFloat("turnover", s.Turnover),
			YtdChangeRate:         p.parseFloat("ytd_change_rate", s.YtdChangeRate),
			TurnoverRate:          p.parseFloat("turnover_rate", s.TurnoverRate),
			TotalMarketValue:      p.parseFloat("total_market_value", s.TotalMarketValue),
			CapitalFlow:           p.parseFloat("capital_flow", s.CapitalFlow),
			Amplitude:             p.parseFloat("amplitude", s.Amplitude),
			VolumeRatio:           p.parseFloat("volume_ratio", s.VolumeRatio),
			PeTtmRatio:            p.parseFloat("pe_ttm_ratio", s.PeTtmRatio),
			PbRatio:               p.parseFloat("pb_ratio", s.PbRatio),
			DividendRatioTtm:      p.parseFloat("dividend_ratio_ttm", s.DividendRatioTtm),
			FiveDayChangeRate:     p.parseFloat("5d_change_rate", s.FiveDayChangeRate),
			TenDayChangeRate:      p.parseFloat("10d_change_rate", s.TenDayChangeRate),
			HalfYearChangeRate:    p.parseFloat("hy_change_rate", s.HalfYearChangeRate),
			FiveMinutesChangeRate: p.parseFloat("5m_change_rate", s.FiveMinutesChangeRate),
			ExpiryDate:            s.ExpiryDate,
			StrikePrice:           p.parseFloat("strike_price", s.StrikePrice),
			UpperStrikePrice:      p.parseFloat("upper_strike_price", s.UpperStrikePrice),
			LowerStrikePrice:      p.parseFloat("lower_strike_price", s.LowerStrikePrice),
			OutstandingQty:        s.OutstandingQty,
			OutstandingRatio:      p.parseFloat("outstanding_ratio", s.OutstandingRatio),
			Premium:               p.parseFloat("premium", s.Premium),
			ItmOtm:                p.parseFloat("item_otm", s.ItmOtm),
			ImpliedVolatility:     p.parseFloat("implied_volatility", s.ImpliedVolatility),
			WarrantDelta:          p.parseFloat("warrant_delta", s.WarrantDelta),
			CallPrice:             p.parseFloat("call_price", s.CallPrice),
			ToCallPrice:           p.parseFloat("to_call_Price", s.ToCallPrice),
			EffectiveLeverage:     p.parseFloat("effective_leverage", s.EffectiveLeverage),
			LeverageRatio:         p.parseFloat("leverage_ratio", s.LeverageRatio),
			ConversionRatio:       p.parseFloat("conversion_ratio", s.ConversionRatio),
			BalancePoint:          p.parseFloat("balance_point", s.BalancePoint),
			OpenInterest:          s.OpenInterest,
			Delta:                 p.parseFloat("delta", s.Delta),
			Gamma:                 p.parseFloat("gamma", s.Gamma),
			Theta:                 p.parseFloat("theta", s.Theta),
			Vega:                  p.parseFloat("vega", s.Vega),
			Rho:                   p.parseFloat("rho", s.Rho),
		}
		if err := p.Error(); err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}
	return fs, nil
}

func (c *quoteLongConn) handlePushPkg(header *protocol.PushPkgHeader, body []byte, pkgErr error) {
	if pkgErr != nil {
		glog.V(2).Infof("Error getting pushed package: len=%d err=%v", len(body), pkgErr)
		return
	}
	switch protocol.Command(header.CmdCode) {
	case protocol.CmdPushQuoteData:
		if c.OnPushQuote == nil {
			glog.V(3).Infof("Ignore pushed quote: len=%d, err=%v", len(body), pkgErr)
			return
		}
		var resp quote.PushQuote
		if err := proto.Unmarshal(body, &resp); err != nil {
			glog.V(2).Infof("Invalid pushed quote: len=%d, err=%v", len(body), err)
			return
		}
		p := &parser{}
		q := &PushQuote{
			Symbol:          Symbol(resp.Symbol),
			Sequence:        resp.Sequence,
			LastDone:        p.parseFloat("last_done", resp.LastDone),
			Open:            p.parseFloat("open", resp.Open),
			High:            p.parseFloat("high", resp.High),
			Low:             p.parseFloat("low", resp.Low),
			Timestamp:       resp.Timestamp,
			Volume:          resp.Volume,
			Turnover:        p.parseFloat("turnover", resp.Turnover),
			TradeStatus:     resp.TradeStatus,
			TradeSession:    resp.TradeSession,
			CurrentVolume:   resp.CurrentVolume,
			CurrentTurnover: p.parseFloat("current_turnover", resp.CurrentTurnover),
		}
		if err := p.Error(); err != nil {
			glog.V(2).Infof("Error parsing pushed quote: %v", err)
			return
		}
		c.OnPushQuote(q)

	case protocol.CmdPushOrderBookData:
		if c.OnPushOrderBook == nil {
			glog.V(3).Infof("Ignore pushed order book: len=%d err=%v", len(body), pkgErr)
			return
		}
		var resp quote.PushDepth
		if err := proto.Unmarshal(body, &resp); err != nil {
			glog.V(2).Infof("Invalid pushed order book: len=%d, err=%v", len(body), err)
			return
		}
		p := &parser{}
		ob := &PushOrderBook{
			Symbol:   Symbol(resp.Symbol),
			Sequence: resp.Sequence,
		}
		for _, bid := range resp.Bid {
			ob.Bid = append(ob.Bid, &OrderBook{
				Position: bid.Position,
				Price:    p.parseFloat("bid_price", bid.Price),
				Volume:   bid.Volume,
				OrderNum: bid.OrderNum,
			})
			if err := p.Error(); err != nil {
				glog.V(2).Infof("Error parsing pushed order book bid data: %v", err)
				return
			}
		}
		for _, ask := range resp.Ask {
			ob.Ask = append(ob.Ask, &OrderBook{
				Position: ask.Position,
				Price:    p.parseFloat("bid_price", ask.Price),
				Volume:   ask.Volume,
				OrderNum: ask.OrderNum,
			})
			if err := p.Error(); err != nil {
				glog.V(2).Infof("Error parsing pushed order book ask data: %v", err)
				return
			}
		}
		c.OnPushOrderBook(ob)

	case protocol.CmdPushBrokersData:
		if c.OnPushBrokers == nil {
			glog.V(3).Infof("Ignore pushed brokers: len=%d err=%v", len(body), pkgErr)
			return
		}
		var resp quote.PushBrokers
		if err := proto.Unmarshal(body, &resp); err != nil {
			glog.V(2).Infof("Invalid pushed brokers: len=%d, err=%v", len(body), err)
			return
		}
		b := &PushBrokers{
			Symbol:   Symbol(resp.Symbol),
			Sequence: resp.Sequence,
		}
		for _, bid := range resp.BidBrokers {
			b.Bid = append(b.Bid, &Broker{
				Position: bid.Position,
				IDs:      bid.BrokerIds,
			})
		}
		for _, ask := range resp.AskBrokers {
			b.Ask = append(b.Ask, &Broker{
				Position: ask.Position,
				IDs:      ask.BrokerIds,
			})
		}
		c.OnPushBrokers(b)

	case protocol.CmdPushTickersData:
		if c.OnPushTickers == nil {
			glog.V(3).Infof("Ignore pushed tickers: len=%d err=%v", len(body), pkgErr)
			return
		}
		var resp quote.PushTrade
		if err := proto.Unmarshal(body, &resp); err != nil {
			glog.V(2).Infof("Invalid pushed tickers: len=%d, err=%v", len(body), err)
			return
		}
		t := &PushTickers{
			Symbol:   Symbol(resp.Symbol),
			Sequence: resp.Sequence,
		}
		p := &parser{}
		for _, ticker := range resp.Trade {
			t.Tickers = append(t.Tickers, &Ticker{
				Price:        p.parseFloat("price", ticker.Price),
				Volume:       ticker.Volume,
				Timestamp:    ticker.Timestamp,
				TradeType:    ticker.TradeType,
				Dir:          TradeDir(ticker.Direction),
				TradeSession: ticker.TradeSession,
			})
			if err := p.Error(); err != nil {
				glog.V(2).Infof("Error parsing ticker data: %v", err)
			}
		}
		c.OnPushTickers(t)
	}
}

type QuoteClient struct {
	*client
	*quoteLongConn
}

func NewQuoteClient(conf *Config) (*QuoteClient, error) {
	c, err := newClient(conf)
	if err != nil {
		return nil, err
	}
	qc := newQuoteLongConn(c.config.QuoteEndpoint, c)
	qc.reconnectInterval = time.Duration(c.config.ReconnectInterval) * time.Second
	qc.Enable(true)
	return &QuoteClient{client: c, quoteLongConn: qc}, nil
}

type watchedGroupsResp struct {
	Code int `json:"code"`
	Data struct {
		Groups []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Securities []struct {
				Symbol       string `json:"symbol"`
				Market       string `json:"market"`
				Name         string `json:"name"`
				WatchedPrice string `json:"watched_price"`
				WatchedAt    string `json:"watched_at"`
			} `json:"securities"`
		} `json:"groups"`
	} `json:"data"`
}

func (c *QuoteClient) GetWatchedGroups() ([]*WatchedGroup, error) {
	var resp watchedGroupsResp
	var groups []*WatchedGroup
	if err := c.request(httpGET, watchedGroupURLPath, nil, nil, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("response %+v is not successful", resp)
	}
	p := &parser{}
	log.Printf("%v\n", resp.Data.Groups)
	for _, g := range resp.Data.Groups {
		wg := &WatchedGroup{
			ID:   g.ID,
			Name: g.Name,
		}
		for _, s := range g.Securities {
			wg.Securities = append(wg.Securities, &WatchedSecurity{
				Symbol:       MakeSymbol(s.Symbol, Market(s.Market)),
				Name:         s.Name,
				WatchedPrice: p.parseFloat("watched_price", s.WatchedPrice),
				WatchedAt:    p.parseInt("watched_at", s.WatchedAt),
			})
		}
		groups = append(groups, wg)
	}
	return groups, p.Error()
}

func (c *QuoteClient) Close() {
	c.quoteLongConn.Enable(false)
	c.client.Close()
}
