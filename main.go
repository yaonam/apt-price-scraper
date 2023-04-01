package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/net/html"
)

// Edit these vars as necessary -----------------------------------------------------
var LeftLimit = time.Date(2023, time.June, 1, 0, 0, 0, 0, time.Local)
var RightLimit = time.Date(2023, time.July, 10, 0, 0, 0, 0, time.Local)
var MaxLeaseDuration = 3 // 5+x months
// ---------------------------------------------------------------------------------

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

type DiscordMessage struct {
	Content  string `json:"content"`
	UserName string `json:"username"`
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
	var allApartments Apartments
	if jsonErr := json.Unmarshal(body, &allApartments); jsonErr != nil {
		log.Fatal(jsonErr)
	}
	log.Printf("Fetched %v apartments", len(allApartments))

	// Filter by studio and move-in date
	filteredAparments := allApartments.filter()

	// Calculate pricing (best?), html tokenizer
	filteredAparments.populateQuote()

	// Show pricing (ordered?)
	prettyApts, prettyErr := json.MarshalIndent(filteredAparments, "", "  ")
	if prettyErr != nil {
		log.Fatal(prettyErr)
	}
	log.Print(string(prettyApts))

	// Webhook stuff
	godotenv.Load()
	var WebhookURL string = os.Getenv("WEBHOOK_URL")
	content := "```json\n" + string(prettyApts) + "\n```"
	discordMessage, _ := json.Marshal(DiscordMessage{Content: content, UserName: "Apt Price Scraper"})
	webhookReq, _ := http.NewRequest("POST", WebhookURL, bytes.NewBuffer(discordMessage))
	webhookReq.Header.Set("Content-Type", "application/json")
	_, webhookErr := http.DefaultClient.Do(webhookReq)
	if webhookErr != nil {
		log.Fatal(webhookErr)
	}
}

func (apartments *Apartments) filter() Apartments {
	// Filter by studio
	var studioApartments Apartments
	for _, apartment := range *apartments {
		if apartment.Bedroom == "Studio" {
			studioApartments = append(studioApartments, apartment)
		}
	}
	log.Printf("Filtered %v studio apartments", len(studioApartments))
	// Filter by move-in date
	var juneJulyStudioApts Apartments
	for _, apartment := range studioApartments {
		availability := apartment.Availability
		if availability.After(LeftLimit) && availability.Before(RightLimit) {
			juneJulyStudioApts = append(juneJulyStudioApts, apartment)
		}
	}
	log.Printf("Filtered %v apartments by move-in date", len(juneJulyStudioApts))
	return juneJulyStudioApts
}

func (apartments *Apartments) populateQuote() {
	for i, apartment := range *apartments {
		// Get price matrix
		req, _ := http.NewRequest("GET", PricingMatrixURL, nil)
		req.URL.RawQuery = url.Values{
			"contentclass":      {"pricingmatrix"},
			"UnitId":            {apartment.IDValue},
			"UnitAvailableDate": {apartment.Availability.Format("01/02/2006")},
		}.Encode()
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		resp, respErr := http.DefaultClient.Do(req)
		if respErr != nil {
			log.Fatal(respErr)
		}
		tokenizer := html.NewTokenizer(resp.Body)
		defer resp.Body.Close()
		bestQuote := apartment.getBestQuote(tokenizer)
		resp.Body.Close()
		(*apartments)[i].Quote = bestQuote
	}
}

func (apartment *Apartment) getBestQuote(tokenizer *html.Tokenizer) float64 {
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
		// Lease duration of < (5+x) months
		if i < MaxLeaseDuration && quote < bestQuote {
			bestQuote = quote
		}
		// Early move-out, assume after 6 months
		earlyMoveOutQuote := quote * 8 / 7
		if earlyMoveOutQuote < bestQuote {
			bestQuote = earlyMoveOutQuote
		}
	}
	return bestQuote
}

func getRowFirstQuote(tokenizer *html.Tokenizer, rowName string) string {
	// Skip to row in price matrix
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			log.Print("Error!")
			return ""
		}
		if tokenType == html.StartTagToken {
			_, val, _ := tokenizer.TagAttr()
			if string(val) == rowName {
				break
			}
		}
	}
	// Find first quote
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			log.Print("Error!")
			return ""
		}
		text := tokenizer.Text()
		if tokenType == html.TextToken && text[0] == '$' {
			newQuote := string(text)
			// log.Printf("%v: %v", rowName, newQuote)
			return newQuote
		}
	}
}
