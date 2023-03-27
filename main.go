package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var MoveInDate time.Time = time.Date(2023, time.July, 1, 0, 0, 0, 0, time.Local)
var leftLimit = time.Date(2023, time.June, 1, 0, 0, 0, 0, time.Local)
var july = time.Date(2023, time.July, 10, 0, 0, 0, 0, time.Local)

const AvailableUnitsURL string = "https://protokendallsq.com/floorplans/"
const PricingMatrixURL string = "https://protokendallsq.securecafe.com/rcloadcontent.ashx"

type Apartment struct {
	ID           string    `json:"id"`
	Unit         string    `json:"unit"`
	Bedroom      string    `json:"bedroom"`
	SqFt         uint      `json:"sq_ft"`
	MinRent      uint      `json:"min_rent"`
	MaxRent      uint      `json:"max_rent"`
	Availability time.Time `json:"availability"`
	Floor        string    `json:"floor"`
	Quote        uint      `json:"quote"`
}

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
	var allApartments []Apartment
	jsonErr := json.Unmarshal(body, &allApartments)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	log.Printf("Fetched %v apartments", len(allApartments))

	// Filter by studio
	var studioApartments []Apartment
	for _, apartment := range allApartments {
		if apartment.Bedroom == "Studio" {
			studioApartments = append(studioApartments, apartment)
		}
	}
	// Filter by move-in date
	var juneJulyStudioApts []Apartment
	for _, apartment := range studioApartments {
		availability := apartment.Availability
		if availability.After(leftLimit) && availability.Before(july) {
			juneJulyStudioApts = append(juneJulyStudioApts, apartment)
		}
	}
	log.Printf("Filtered %v apartments", len(juneJulyStudioApts))

	// Calculate pricing (best?), html tokenizer
	juneJulyStudioApts[0].populateQuote(MoveInDate)

	// Show pricing (ordered?)
	prettyApts, prettyErr := json.MarshalIndent(juneJulyStudioApts, "", "  ")
	if prettyErr != nil {
		log.Fatal(prettyErr)
	}
	log.Print(string(prettyApts))
}

func (apartment *Apartment) populateQuote(moveInDate time.Time) {
	// Get price matrix
	req, _ := http.NewRequest("GET", PricingMatrixURL, nil)
	req.URL.RawQuery = url.Values{
		"contentclass":      {"pricingmatrix"},
		"UnitId":            {"30653186"},
		"UnitAvailableDate": {"5/11/2023"},
	}.Encode()
	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		log.Fatal(respErr)
	}
	// body, readErr := io.ReadAll(resp.Body)
	tokenizer := html.NewTokenizer(resp.Body)
	defer resp.Body.Close()
	// log.Print(string(body))
	for {
		tokenizer.Next()
		log.Printf("Token: %v", tokenizer.Token().String())
		log.Printf("Matched?: %v", tokenizer.Token().String() == "<tr id=\"Pricerow11\">")
		if tokenizer.Token().String() == "<tr id=\"Pricerow0\">" {
			log.Printf("Token: %v", tokenizer.Token())
			break
		} else if tokenizer.Token().Type == html.ErrorToken {
			log.Print("Error!")
			log.Print("\"")
			break
		}
	}
}
