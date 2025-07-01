package types

import (
	"fmt"
	"time"
)

// Account represents a trading account
type Account struct {
	ID          string
	Type        string
	Name        string
	Status      string
	OpenedDate  time.Time
	ClosedDate  time.Time
	AccessLevel string
}

// Instrument represents a trading instrument
type Instrument struct {
	FIGI                  string
	Ticker                string
	ClassCode             string
	ISIN                  string
	Lot                   int32
	Currency              string
	Name                  string
	Exchange              string
	CountryOfRisk         string
	CountryOfRiskName     string
	InstrumentType        string
	TradingStatus         string
	OTCFlag               bool
	BuyAvailableFlag      bool
	SellAvailableFlag     bool
	MinPriceIncrement     *MoneyValue
	APITradeAvailableFlag bool
}

// MoneyValue represents a monetary value with currency
type MoneyValue struct {
	Currency string
	Units    int64
	Nano     int32
}

// ToFloat converts MoneyValue to float64
func (m *MoneyValue) ToFloat() float64 {
	return float64(m.Units) + float64(m.Nano)/1e9
}

// NewMoneyValue creates MoneyValue from float64
func NewMoneyValue(value float64, currency string) *MoneyValue {
	units := int64(value)
	nano := int32((value - float64(units)) * 1e9)

	return &MoneyValue{
		Currency: currency,
		Units:    units,
		Nano:     nano,
	}
}

// String returns string representation of MoneyValue
func (m *MoneyValue) String() string {
	return fmt.Sprintf("%.2f %s", m.ToFloat(), m.Currency)
}

// Quotation represents price quotation
type Quotation struct {
	Units int64
	Nano  int32
}

// ToFloat converts Quotation to float64
func (q *Quotation) ToFloat() float64 {
	return float64(q.Units) + float64(q.Nano)/1e9
}

// NewQuotation creates Quotation from float64
func NewQuotation(value float64) *Quotation {
	units := int64(value)
	nano := int32((value - float64(units)) * 1e9)

	return &Quotation{
		Units: units,
		Nano:  nano,
	}
}

// String returns string representation of Quotation
func (q *Quotation) String() string {
	return fmt.Sprintf("%.4f", q.ToFloat())
}

// OrderDirection represents order direction
type OrderDirection int32

const (
	OrderDirectionUnspecified OrderDirection = 0
	OrderDirectionBuy         OrderDirection = 1
	OrderDirectionSell        OrderDirection = 2
)

// OrderType represents order type
type OrderType int32

const (
	OrderTypeUnspecified OrderType = 0
	OrderTypeLimit       OrderType = 1
	OrderTypeMarket      OrderType = 2
	OrderTypeBestPrice   OrderType = 3
)

// OrderState represents order state
type OrderState int32

const (
	OrderStateUnspecified   OrderState = 0
	OrderStateFill          OrderState = 1
	OrderStateRejected      OrderState = 2
	OrderStateCancelled     OrderState = 3
	OrderStateNew           OrderState = 4
	OrderStatePartiallyFill OrderState = 5
)

// CandleInterval represents candle time interval
type CandleInterval int32

const (
	CandleIntervalUnspecified CandleInterval = 0
	CandleInterval1Min        CandleInterval = 1
	CandleInterval5Min        CandleInterval = 2
	CandleInterval15Min       CandleInterval = 3
	CandleInterval1Hour       CandleInterval = 4
	CandleInterval1Day        CandleInterval = 5
	CandleInterval2Min        CandleInterval = 6
	CandleInterval3Min        CandleInterval = 7
	CandleInterval10Min       CandleInterval = 8
	CandleInterval30Min       CandleInterval = 9
	CandleInterval2Hour       CandleInterval = 10
	CandleInterval4Hour       CandleInterval = 11
	CandleInterval1Week       CandleInterval = 12
	CandleInterval1Month      CandleInterval = 13
)

// Candle represents OHLCV candle data
type Candle struct {
	FIGI       string
	Interval   CandleInterval
	Open       *Quotation
	High       *Quotation
	Low        *Quotation
	Close      *Quotation
	Volume     int64
	Time       time.Time
	IsComplete bool
}

// Trade represents a trade
type Trade struct {
	FIGI      string
	Direction OrderDirection
	Price     *Quotation
	Quantity  int64
	Time      time.Time
}

// OrderBook represents order book (market depth)
type OrderBook struct {
	FIGI      string
	Depth     int32
	Bids      []*Order
	Asks      []*Order
	Time      time.Time
	LimitUp   *Quotation
	LimitDown *Quotation
}

// Order represents an order in order book or trading order
type Order struct {
	Price    *Quotation
	Quantity int64
}

// Position represents a position in portfolio
type Position struct {
	FIGI                     string
	InstrumentType           string
	Quantity                 *Quotation
	AveragePositionPrice     *MoneyValue
	ExpectedYield            *Quotation
	CurrentNKD               *MoneyValue
	AveragePositionPricePt   *Quotation
	CurrentPrice             *Quotation
	AveragePositionPriceFifo *MoneyValue
	QuantityLots             *Quotation
}

// Portfolio represents account portfolio
type Portfolio struct {
	TotalAmountShares     *MoneyValue
	TotalAmountBonds      *MoneyValue
	TotalAmountEtf        *MoneyValue
	TotalAmountCurrencies *MoneyValue
	TotalAmountFutures    *MoneyValue
	ExpectedYield         *Quotation
	Positions             []*Position
}

// Error represents API error
type Error struct {
	Code    string
	Message string
	Details string
}

// Error implements error interface
func (e *Error) Error() string {
	return fmt.Sprintf("tinkoff api error: %s - %s", e.Code, e.Message)
}
