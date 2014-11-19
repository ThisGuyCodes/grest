package grest

import (
	"errors"
	"time"
)

const (
	referenceTime = "2006-01-02T15:04:05"
)

var (
	RateExceeded = errors.New("Crest per-ip rate exceeded.")
	DoesntExist  = errors.New("Item or Region doesn't exist.")
)

// We need a custom Time type because the dates that come from crest don't have
// a 'Z' suffix like in the standard.
type CrestTime struct {
	time.Time
}

func (t *CrestTime) UnmarshalJSON(data []byte) error {
	// time.Time expects a 'Z' suffix, with no location, for UTC time.
	newData := make([]byte, len(data)+1)

	// If we simply did append(data, 'Z') then the allocation for that slice
	// would be nearly double what it needs, this makes it precise.
	copy(newData, data)
	// We're working with a quoted string, so yea.
	newData[len(newData)-2], newData[len(newData)-1] = 'Z', '"'
	return t.Time.UnmarshalJSON(newData)
}

func (t *CrestTime) MarshalJSON() ([]byte, error) {
	// Here we make sure to use UTC time and use the built in marshal.
	data, err := t.Time.In(time.UTC).MarshalJSON()

	// data could be nil if there was an error.
	if len(data) > 0 {
		// Then remove the 'Z' suffix.
		// Like above, this is a quoted string, so we need to leave it quoted.
		data = data[:len(data)-1]
		data[len(data)] = '"'
	}
	return data, err
}

// This is an abstract market item.
// This is never represented in the crest API, but instead represents the values
// used to retrieve data.
type Item struct {
	Id       int
	Name     string
	OnMarket bool
}

// This is an abstract region item.
// This is never represented in the crest API, but instead represents the values
// used to retrieve data.
type Region struct {
	Id        int
	Name      string
	HasMarket bool
}

type ProcessedHistory struct {
	// This is each of the market days, ordered by date.
	Days      []*ItemMarketDay
	StartDate time.Time
	EndDate   time.Time
}

type ItemMarketDay struct {
	LowPrice   float64   `json:"lowPrice"`
	HighPrice  float64   `json:"highPrice"`
	AvgPrice   float64   `json:"avgPrice"`
	OrderCount int64     `json:"orderCount"`
	Volume     int64     `json:"volume"`
	Date       CrestTime `json:"date"`
	// There's also volume_str, and orderCount_str, but those are useless
}

type ItemMarketHistory struct {
	Days       []*ItemMarketDay `json:"items"`
	PageCount  int              `json:"pageCount"`
	TotalCount int              `json:"totalCount"`
	// There's also totalCount_str, and pageCount_str, but those are useless
}

type CrestError struct {
	Message       string `json:"message"`
	Key           string `json:"key"`
	Exceptiontype string `json:"exceptionType"`
}
