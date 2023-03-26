package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const ProtoURL string = "https://protokendallsq.com/floorplans/"

func main() {
	// Get all available apts
	availableUnitsData := url.Values{
		"action": {"available-units"},
	}
	req, _ := http.NewRequest("POST", ProtoURL, strings.NewReader(availableUnitsData.Encode()))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := http.DefaultClient.Do(req)
	// resp, err := http.PostForm(ProtoURL, availableUnitsData)
	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		resp.Body.Close()
		fmt.Println(string(body))
		// var res map[string]interface{}
		// json.NewDecoder(resp.Body).Decode(&res)
		// fmt.Println(res)
		// fmt.Println(res["json"])
	}
	// Filter by studio
	// Filter by move-in date
	// Calculate pricing (best?), html tokenizer
	// Show pricing (ordered?)
}
