package grest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	urlFormat            = "https://public-crest.eveonline.com/market/%d/types/%d/history/"
	notImplimentedString = "Error %d is not implimented."
)

// Get a market history for the item and a given region
func (i *Item) GetHistoryForRegionId(id int) (*ItemMarketHistory, error) {
	url := fmt.Sprintf(urlFormat, id, i.Id)
	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(response.Body)

	if response.StatusCode == 200 {
		// If success
		history := new(ItemMarketHistory)
		err = decoder.Decode(&history)
		return history, err
	} else if response.StatusCode == 503 {
		// Rate exceeded error
		return nil, RateExceeded
	} else if response.StatusCode == 404 {
		// Doesn't exist
		return nil, DoesntExist
	} else {
		// What?
		panic(fmt.Sprintf(notImplimentedString, response.StatusCode))
	}
}

// Takes in a market history, and produces a processed history.
// Because of the way this is processed, and any 'missing' days
// will likely (not tested) result in nil indices.
// This is O(4n), one each for: before, iszero, after, assignment
func (history ItemMarketHistory) Process() ProcessedHistory {
	processed := ProcessedHistory{}

	// Get the start and end date
	for _, day := range history.Days {
		if day.Date.Before(processed.StartDate) || processed.StartDate.IsZero() {
			processed.StartDate = day.Date.Time
		} else if day.Date.After(processed.EndDate) {
			processed.EndDate = day.Date.Time
		}
	}

	// Allocate the days
	delta := int(processed.EndDate.Sub(processed.StartDate) / time.Hour / 24)
	processed.Days = make([]*ItemMarketDay, delta+1)

	// Place the days
	for _, day := range history.Days {
		// n = time since the earliest date in days
		n := day.Date.Sub(processed.StartDate) / time.Hour / 24
		processed.Days[n] = day
	}

	return processed
}

// Helper function for loading items from a file (or other reader)
func LoadItems(file io.Reader) ([]Item, error) {
	decoder := json.NewDecoder(file)
	var items []Item
	err := decoder.Decode(&items)
	return items, err
}

// Allows for cleaner syntax
func MustLoadItems(file io.Reader) []Item {
	items, err := LoadItems(file)
	if err != nil {
		panic(fmt.Sprintf("LoadItems returned an error:", err))
	}
	closeReader(file)
	return items
}

// Helper function for loading regions from a file (or other reader)
func LoadRegions(file io.Reader) ([]Region, error) {
	decoder := json.NewDecoder(file)
	var regions []Region
	err := decoder.Decode(&regions)
	return regions, err
}

// Allows for cleaner syntax
func MustLoadRegions(file io.Reader) []Region {
	regions, err := LoadRegions(file)
	if err != nil {
		panic(fmt.Sprintf("LoadRegions returned an error:", err))
	}
	closeReader(file)
	return regions
}

func closeReader(file io.Reader) {
	if closer, ok := file.(io.ReadCloser); ok {
		closer.Close()
	}
}
