// Weather CLI
// Marc Riart, 20240219

// Usage: weacli [-a] [-f] [-h] [-r] [-x] city1 city2...
// Credits to api.open-meteo.com for the meteo.
// Credits to openstreetmap for geocoding.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Weather struct {
	Latitude  float64
	Longitude float64
	Address   string
	Current   struct {
		Temperature_2m       float64 `json:"temperature_2m"`
		Relative_humidity_2m int     `json:"relative_humidity_2m"`
		Rain                 float64 `json:"rain"`
		Weather_code         int     `json:"weather_code"`
	} `json:"current"`
}

// Formula: Fahrenheit = (Celsius * 9/5) + 32
func celsiusToFahrenheit(celsius float64) float64 {

	return celsius*(9.0/5.0) + 32.0
}

// Function that gets all weather paramenters.
// It fills the weathers map by adding the parameters of the city.
func getWeather(wg *sync.WaitGroup, weathers map[string]Weather, city string) {

	defer wg.Done()
	w := Weather{}
	l := Location{}

	// Get latitud and longitud to compose url.
	l.GetLocation(city)
	w.Latitude, _ = strconv.ParseFloat(l.Lat, 64)
	w.Longitude, _ = strconv.ParseFloat(l.Lon, 64)
	w.Address = l.Displayname

	// Get weather conditions from open-meteo, unmarshal and put in the variale w Weather.
	url := "https://api.open-meteo.com/v1/forecast?latitude="
	url += l.Lat + "&longitude=" + l.Lon
	url += "&current=temperature_2m,relative_humidity_2m,rain,weather_code"
	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error when accessing the open-meteo web api:", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	errUnmarshal := json.Unmarshal(body, &w)
	if errUnmarshal != nil {
		fmt.Println("Error when unmarshalling after reception of open-meteo:", errUnmarshal)
	}

	// Add to weather for the city.
	weathers[city] = w
}

// Print the results acording to the flags.
func printWeathers(weathers map[string]Weather, a, f, h, r, x *bool) {

	for city, w := range weathers {

		fmt.Printf("City: %s\n", city)

		if *f {
			fmt.Printf("Temperature: %.2fF\n", celsiusToFahrenheit(w.Current.Temperature_2m))
		} else {
			fmt.Printf("Temperature: %.2fC\n", w.Current.Temperature_2m)
		}

		fmt.Printf("Weather code: %d", w.Current.Weather_code)
		switch w.Current.Weather_code {
		case 0:
			fmt.Printf(" - Clear sky\n")
		case 1, 2, 3:
			fmt.Printf(" - Mainly clear, partly cloudy, and overcast\n")
		case 45, 48:
			fmt.Printf(" - Fog and depositing rime fog\n")
		case 51, 53, 55:
			fmt.Printf(" - Drizzle: Light, moderate, and dense intensity\n")
		case 56, 57:
			fmt.Printf(" - Freezing Drizzle: Light and dense intensity\n")
		case 61, 63, 65:
			fmt.Printf(" - Rain: Slight, moderate and heavy intensity\n")
		case 66, 67:
			fmt.Printf(" - Freezing Rain: Light and heavy intensity\n")
		case 71, 73, 75:
			fmt.Printf(" - Snow fall: Slight, moderate, and heavy intensity\n")
		case 77:
			fmt.Printf(" - Snow grains\n")
		case 80, 81, 82:
			fmt.Printf(" - Rain showers: Slight, moderate, and violent\n")
		case 85, 86:
			fmt.Printf(" - Snow showers slight and heavy\n")
		case 95:
			fmt.Printf(" - Thunderstorm: Slight or moderate\n")
		case 96, 99:
			fmt.Printf(" - Thunderstorm with slight and heavy hail\n")
		}

		if *a {
			*x, *h, *r = true, true, true
		}

		if *x {
			fmt.Printf("Lat, lon, address: %f, %f, %s\n", w.Latitude, w.Longitude, w.Address)
		}

		if *h {
			fmt.Printf("Humidity: %d%%\n", w.Current.Relative_humidity_2m)
		}

		if *r {
			fmt.Printf("Rain (preceding hour): %.2fmm\n", w.Current.Rain)
		}

		fmt.Printf("\n")
	}
}

// Main func: get arguments, fill Weather struct(s), and print in loop the range.
func main() {

	// Get flags.
	a := flag.Bool("a", false, "All the options together (except fahrenheti)")
	f := flag.Bool("f", false, "If you want fahrenheit")
	h := flag.Bool("h", false, "Display relative humidity")
	r := flag.Bool("r", false, "Display the current rain")
	x := flag.Bool("x", false, "If you want to explain the city adress (Barcelona Spain or Italy)")

	flag.Parse()

	// Get list of arguments (cities). Place in slice cities.
	cities := flag.Args()
	if len(cities) < 1 {
		fmt.Println(`Usage: weacli [-a] [-f] [-h] [-r] city1 city2...
	-a    All the options together (except fahrenheti)
	-f    If you want fahrenheit
	-h    Display relative humidity
	-r    Display the current rain
	-x    If you want to explain the city adress (Barcelona Spain or South America)`)
		os.Exit(1)
	}

	// Get the list of coordinates and weather conditions for all cities in arguments.
	// Place results in a map[city]=Weater_struct.
	weathers := map[string]Weather{}
	var wg sync.WaitGroup

	for _, city := range cities {
		wg.Add(1)
		go getWeather(&wg, weathers, city)
	}
	wg.Wait()

	printWeathers(weathers, a, f, h, r, x)
}
