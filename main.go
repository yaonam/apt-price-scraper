package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const AvailableUnitsURL string = "https://protokendallsq.com/floorplans/"
const PricingMatrixURL string = "https://protokendallsq.securecafe.com/rcloadcontent.ashx"

type Apartment struct {
	ID           string    `json:"id"`
	Unit         string    `json:"unit"`
	Bedroom      string    `json:"bedroom"`
	SqFt         uint      `json:"sq_ft"`
	MinRent      uint      `json:"min_rent"`
	MaxRent      uint      `json:"max_rent"`
	Availability time.Time `json:"availability"` // "2022-06-08T00:00:00-05:00"
	Floor        string    `json:"floor"`
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

	// Filter by studio
	var studioApartments []Apartment
	for _, apartment := range allApartments {
		if apartment.Bedroom == "Studio" {
			studioApartments = append(studioApartments, apartment)
		}
	}

	// Filter by move-in date
	var juneJulyStudioApts []Apartment
	june := time.Date(2023, time.June, 1, 0, 0, 0, 0, time.Local)
	july := time.Date(2023, time.July, 10, 0, 0, 0, 0, time.Local)
	for _, apartment := range studioApartments {
		availability := apartment.Availability
		if availability.After(june) && availability.Before(july) {
			juneJulyStudioApts = append(juneJulyStudioApts, apartment)
		}
	}

	// Calculate pricing (best?), html tokenizer

	// Show pricing (ordered?)
	prettyApts, prettyErr := json.MarshalIndent(juneJulyStudioApts, "", "  ")
	if prettyErr != nil {
		log.Fatal(prettyErr)
	}
	log.Print(string(prettyApts))
}

// func (apartment *Apartment) populateQuote(moveInDate time.Time) {
// 	// Get price matrix
// 	availableUnitsData := url.Values{
// 		"action": {"available-units"},
// 	}
// 	req, _ := http.NewRequest("POST", AvailableUnitsURL, strings.NewReader(availableUnitsData.Encode()))
// 	req.Header.Set("X-Requested-With", "XMLHttpRequest")
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
// 	resp, respErr := http.DefaultClient.Do(req)
// 	if respErr != nil {
// 		log.Fatal(respErr)
// 	}
// 	body, readErr := io.ReadAll(resp.Body)
// 	if readErr != nil {
// 		log.Println(readErr)
// 	}
// 	resp.Body.Close()
// }
