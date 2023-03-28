package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Edit these vars as necessary -----------------------------------------------------
var MoveInDate time.Time = time.Date(2023, time.July, 1, 0, 0, 0, 0, time.Local)
var leftLimit = time.Date(2023, time.June, 1, 0, 0, 0, 0, time.Local)
var rightLimit = time.Date(2023, time.July, 10, 0, 0, 0, 0, time.Local)

const AvailableUnitsURL string = "https://protokendallsq.com/floorplans/"
const PricingMatrixURL string = "https://protokendallsq.securecafe.com/rcloadcontent.ashx"

type Apartment struct {
	ID           string    `json:"id"`
	Unit         string    `json:"unit"`
	IDValue      string    `json:"id_value"`
	Bedroom      string    `json:"bedroom"`
	SqFt         uint      `json:"sq_ft"`
	MinRent      uint      `json:"min_rent"`
	MaxRent      uint      `json:"max_rent"`
	Availability time.Time `json:"availability"`
	Floor        string    `json:"floor"`
	Quote        float64   `json:"quote"`
}

type Apartments []Apartment

func main() {
	// Get all available apts
	availableUnitsData := url.Values{
		"action": {"available-units"},
	}
	req, _ := http.NewRequest("POST", AvailableUnitsURL, strings.NewReader(availableUnitsData.Encode()))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		log.Fatal(respErr)
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Println(readErr)
	}
	resp.Body.Close()
	var allApartments Apartments
	jsonErr := json.Unmarshal(body, &allApartments)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	log.Printf("Fetched %v apartments", len(allApartments))

	// Filter by studio
	var studioApartments Apartments
	for _, apartment := range allApartments {
		if apartment.Bedroom == "Studio" {
			studioApartments = append(studioApartments, apartment)
		}
	}
	log.Printf("Filtered %v studio apartments", len(studioApartments))
	// Filter by move-in date
	var juneJulyStudioApts Apartments
	for _, apartment := range studioApartments {
		availability := apartment.Availability
		if availability.After(leftLimit) && availability.Before(rightLimit) {
			juneJulyStudioApts = append(juneJulyStudioApts, apartment)
		}
	}
	log.Printf("Filtered %v apartments by move-in date", len(juneJulyStudioApts))

	// Calculate pricing (best?), html tokenizer
	juneJulyStudioApts.populateQuote(MoveInDate)

	// Show pricing (ordered?)
	prettyApts, prettyErr := json.MarshalIndent(juneJulyStudioApts, "", "  ")
	if prettyErr != nil {
		log.Fatal(prettyErr)
	}
	log.Print(string(prettyApts))
}

func (apartments *Apartments) populateQuote(moveInDate time.Time) {
	for i, apartment := range *apartments {
		// Get price matrix
		req, _ := http.NewRequest("GET", PricingMatrixURL, nil)
		req.URL.RawQuery = url.Values{
			"contentclass":      {"pricingmatrix"},
			"UnitId":            {apartment.IDValue},
			"UnitAvailableDate": {"7/1/2023"},
		}.Encode()
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		resp, respErr := http.DefaultClient.Do(req)
		if respErr != nil {
			log.Fatal(respErr)
		}
		tokenizer := html.NewTokenizer(resp.Body)
		defer resp.Body.Close()
		// log.Print(string(body))
		bestQuote := apartment.getBestQuote(tokenizer)
		// log.Printf("Best quote for unit %v: %v", apartment.Unit, bestQuote)
		(*apartments)[i].Quote = bestQuote
	}
}

func (apartment *Apartment) getBestQuote(tokenizer *html.Tokenizer) float64 {
	// TODO check early cancel pricing
	quotes := make([]float64, 13)
	for i := range quotes {
		rowName := "Pricerow" + fmt.Sprint(i)
		rawQuote := getRowFirstQuote(tokenizer, rowName)
		rawQuote = strings.Replace(rawQuote, "$", "", -1)
		rawQuote = strings.Replace(rawQuote, ",", "", -1)
		quotes[i], _ = strconv.ParseFloat(rawQuote, 64)
	}

	bestQuote := quotes[0]
	for i := 0; i < len(quotes); i++ {
		quote := quotes[i]
		// Lease duration of 6-8 months
		if i < 3 && quote < bestQuote {
			bestQuote = quote
		}
		// Early move-out, assume after 6 months
		earlyMoveOutQuote := quote * 7 / 6
		if earlyMoveOutQuote < bestQuote {
			bestQuote = earlyMoveOutQuote
		}
	}
	return bestQuote
}

func getRowFirstQuote(tokenizer *html.Tokenizer, rowName string) string {
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			log.Print("Error!")
			break
		}

		// Look for start tags
		if tokenType == html.StartTagToken {
			_, val, _ := tokenizer.TagAttr()
			// See if id matches row name
			if string(val) == rowName {
				for {
					tokenType := tokenizer.Next()
					if tokenType == html.ErrorToken {
						break
					}
					tagname, _ := tokenizer.TagName()
					if string(tagname) == "label" {
						tokenType := tokenizer.Next() // Get label text
						if tokenType == html.StartTagToken {
							// Skip bold tag
							tokenizer.Next()
						}
						newQuote := string(tokenizer.Text())
						// log.Printf("%v: %v", rowName, newQuote)
						return newQuote
					}
				}
				break
			}
		}
	}
	return ""
}
