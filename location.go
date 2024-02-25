package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Location struct {
	Displayname string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

// Fills the location received as pointer.
func (loc *Location) GetLocation(city string) {
	resp, err := http.DefaultClient.Get("https://nominatim.openstreetmap.org/search?format=json&limit=1&q=" + url.PathEscape(city))
	if err != nil {
		fmt.Println("Error in http.Get of openstreetmap.", err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(body[1:len(body)-1], loc)
	if err != nil {
		fmt.Println("Error in unmarshalling openstreetmap.", err)
	}

	// fmt.Println(loc)
}
