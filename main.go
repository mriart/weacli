// Weather CLI
// Marc Riart, 20240225

// Usage:  weacli [-f days] [options] city1 city2...
// Credits to api.open-meteo.com for the meteo.
// Credits to openstreetmap for geocoding.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const usage = `Usage: weacli [-f days] [-a] [-s] [-h] city1 city2...
	-f <days> The number of forecast days. From 0 to 15. If not specified, current weather (0)
	-a        All weather values, not only the default, temperature and condition
	-s        Show sunrise, sunset and day duration
	-h        Display help`

type Weather struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	Address              string
	Timezone             string `json:"timezone"`
	TimezoneAbbreviation string `json:"timezone_abbreviation"`
	Current              struct {
		Time               string  `json:"time"`
		Temperature2M      float64 `json:"temperature_2m"`
		RelativeHumidity2M int     `json:"relative_humidity_2m"`
		Precipitation      float64 `json:"precipitation"`
		WeatherCode        int     `json:"weather_code"`
		SurfacePressure    float64 `json:"surface_pressure"`
		WindDirection10M   int     `json:"wind_direction_10m"`
		WindGusts10M       float64 `json:"wind_gusts_10m"`
	} `json:"current"`
	Daily struct {
		Time                        []string  `json:"time"`
		WeatherCode                 []int     `json:"weather_code"`
		Temperature2MMax            []float64 `json:"temperature_2m_max"`
		Temperature2MMin            []float64 `json:"temperature_2m_min"`
		Sunrise                     []string  `json:"sunrise"`
		Sunset                      []string  `json:"sunset"`
		DaylightDuration            []float64 `json:"daylight_duration"`
		PrecipitationSum            []float64 `json:"precipitation_sum"`
		PrecipitationHours          []float64 `json:"precipitation_hours"`
		PrecipitationProbabilityMax []int     `json:"precipitation_probability_max"`
		WindGusts10MMax             []float64 `json:"wind_gusts_10m_max"`
		WindDirection10MDominant    []int     `json:"wind_direction_10m_dominant"`
	} `json:"daily"`
}

func main() {
	// Get flags from command line.
	f := flag.Int("f", 0, "The number of forecast days. From 0 to 15. If not specified, current weather (0)")
	a := flag.Bool("a", false, "All weather values, not only the default temperature and condition")
	s := flag.Bool("s", false, "Show sunrise, sunset and day duration")
	h := flag.Bool("h", false, "Display help")

	flag.Parse()

	// If -h, display help and exit.
	if *h {
		fmt.Println(usage)
		return
	}

	// Get the number of days to forecast. Between 0 and 15, default 0.
	forecastDays := 0
	if *f < 0 || *f > 15 {
		*f = 0
	} else {
		forecastDays = *f
	}

	// Get list of arguments (cities). Place in slice cities.
	cities := flag.Args()
	if len(cities) < 1 {
		fmt.Println(usage)
		return
	}

	// Get the list of coordinates and weather conditions for all cities in arguments.
	// Place results in a map[city]=Weater_struct.
	weathers := map[string]Weather{}
	var wg sync.WaitGroup

	for _, city := range cities {
		wg.Add(1)
		go GetWeather(&wg, weathers, city, forecastDays)
	}
	wg.Wait()
	//fmt.Println(weathers)
	PrintWeathers(weathers, a, s)
}

// Function that gets all weather paramenters.
// It fills the weathers map by adding the parameters of the city.
func GetWeather(wg *sync.WaitGroup, weathers map[string]Weather, city string, forecastDays int) {
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
	url += "&current=temperature_2m,relative_humidity_2m,precipitation,weather_code,surface_pressure,wind_direction_10m,wind_gusts_10m"
	url += "&daily=weather_code,temperature_2m_max,temperature_2m_min,sunrise,sunset,daylight_duration,precipitation_sum,precipitation_hours,precipitation_probability_max,wind_gusts_10m_max,wind_direction_10m_dominant"
	url += "&timezone=auto"
	url += "&start_date=" + time.Now().Format("2006-01-02")
	url += "&end_date=" + time.Now().AddDate(0, 0, forecastDays).Format("2006-01-02")
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

// Formula: Fahrenheit = (Celsius * 9/5) + 32
func CelsiusToFahrenheit(celsius float64) float64 {
	return celsius*(9.0/5.0) + 32.0
}

// Print the text explanation of the weather code.
func WeatherCodeExplanation(wc int) string {
	switch wc {
	case 0:
		return "Clear sky"
	case 1, 2, 3:
		return "Mainly clear, partly cloudy, and overcast"
	case 45, 48:
		return "Fog and depositing rime fog"
	case 51, 53, 55:
		return "Drizzle: Light, moderate, and dense intensity"
	case 56, 57:
		return "Freezing Drizzle: Light and dense intensity"
	case 61, 63, 65:
		return "Rain: Slight, moderate and heavy intensity"
	case 66, 67:
		return "Freezing Rain: Light and heavy intensity"
	case 71, 73, 75:
		return "Snow fall: Slight, moderate, and heavy intensity"
	case 77:
		return "Snow grains"
	case 80, 81, 82:
		return "Rain showers: Slight, moderate, and violent"
	case 85, 86:
		return "Snow showers slight and heavy"
	case 95:
		return "Thunderstorm: Slight or moderate"
	case 96, 99:
		return "Thunderstorm with slight and heavy hail"
	default:
		return "Unknown weather code"
	}
}

// Print the results acording to the flags.
func PrintWeathers(weathers map[string]Weather, a, s *bool) {
	for city, w := range weathers {
		// Print the city, location, time, and sunrise/sunset.
		fmt.Printf("City: %s\n", city)
		fmt.Printf("Lat: %f, Lon: %f, Adress: %s\n", w.Latitude, w.Longitude, w.Address)
		fmt.Printf("Time: %s, %s %s\n", w.Current.Time, w.Timezone, w.TimezoneAbbreviation)
		if *s {
			fmt.Printf("Sunrise: %s\n", w.Daily.Sunrise[0])
			fmt.Printf("Sunset: %s\n", w.Daily.Sunset[0])
			fmt.Printf("Dailight duration: %.2fs\n", w.Daily.DaylightDuration[0])
		}

		// Print the current weather conditions.
		fmt.Printf("Now:\n")
		fmt.Printf("  Temperature: %.2f°C\n", w.Current.Temperature2M)
		fmt.Printf("  Weather code: %d - %s\n", w.Current.WeatherCode, WeatherCodeExplanation(w.Current.WeatherCode))
		if *a {
			fmt.Printf("  Precipitation (preceding hour): %.2fmm\n", w.Current.Precipitation)
			fmt.Printf("  Relative humidity: %d%%\n", w.Current.RelativeHumidity2M)
			fmt.Printf("  Preassure: %.2fhPa\n", w.Current.SurfacePressure)
			fmt.Printf("  Wind gusts speed: %.2fKm/h\n", w.Current.WindGusts10M)
			fmt.Printf("  Wind direction (from N): %d°\n", w.Current.WindDirection10M)
		}

		// Print the forecast for the number of provided days.
		for i, v := range w.Daily.Time {
			if i == 0 {
				continue
			}
			if i == 1 {
				fmt.Printf("Tomorrow, %s\n", v)
			} else {
				fmt.Printf("%s, %s\n", v, DayOfWeek(v))
			}
			fmt.Printf("  Weather code: %d - %s\n", w.Daily.WeatherCode[i], WeatherCodeExplanation(w.Daily.WeatherCode[i]))
			fmt.Printf("  Max: %.2f°\n", w.Daily.Temperature2MMax[i])
			fmt.Printf("  Min: %.2f°\n", w.Daily.Temperature2MMin[i])
			if *a {
				fmt.Printf("  Precipitation sum: %.2f\n", w.Daily.PrecipitationSum[i])
				fmt.Printf("  Precipitation probablity: %d%%\n", w.Daily.PrecipitationProbabilityMax[i])
				fmt.Printf("  Precipitation hours: %.2f\n", w.Daily.PrecipitationHours[i])
				fmt.Printf("  Wind gusts speed: %.2fKm/h\n", w.Daily.WindGusts10MMax[i])
				fmt.Printf("  Wind dominnant direction (from N): %d°\n", w.Daily.WindDirection10MDominant[i])
			}
			if *s {
				fmt.Printf("  Sunrise: %s\n", w.Daily.Sunrise[i])
				fmt.Printf("  Sunset: %s\n", w.Daily.Sunset[i])
				fmt.Printf("  Dailight duration: %.2fs\n", w.Daily.DaylightDuration[i])
			}
		}
		fmt.Printf("\n")
	}
}

// Day of week. Given the string 2024-05-14, return Tuesday
func DayOfWeek(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return t.Weekday().String()
}
