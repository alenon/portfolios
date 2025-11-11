package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// Quote represents a stock quote with price and metadata
type Quote struct {
	Symbol          string
	Price           decimal.Decimal
	Open            decimal.Decimal
	High            decimal.Decimal
	Low             decimal.Decimal
	Volume          int64
	PreviousClose   decimal.Decimal
	Change          decimal.Decimal
	ChangePercent   decimal.Decimal
	LastUpdated     time.Time
	MarketCap       *decimal.Decimal
	PE              *decimal.Decimal
	Week52High      *decimal.Decimal
	Week52Low       *decimal.Decimal
	AverageDailyVol *int64
}

// HistoricalPrice represents a historical price point
type HistoricalPrice struct {
	Date     time.Time
	Open     decimal.Decimal
	High     decimal.Decimal
	Low      decimal.Decimal
	Close    decimal.Decimal
	Volume   int64
	AdjClose *decimal.Decimal
}
