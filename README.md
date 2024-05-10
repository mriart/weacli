# Weather CLI
Marc Riart, 20240225

This is a divertimento to learn go. It is a simple CLI for the current weather and n days forecast, in one or more locations.

Usage: weacli [-f days] [-a] [-s] [-h] city1 city2... <br>
	-f days	The number of forecast days. From 0 to 15. If not specified, current weather (0) <br>
	-a		All weather values, not only the default, temperature and condition <br>
	-s		Show sunrise, sunset and day duration <br>
	-h		Display help <br>

Credits to api.open-meteo.com for the meteo. <br>
Credits to openstreetmap for geocoding.