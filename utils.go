package main

import "math"

// Earth's radius in kilometers
const earthRadiusKm = 6371

// rad converts degrees to radians
func rad(deg float64) float64 {
	return deg * math.Pi / 180
}

// haversine calculates the distance between two geographic coordinates in meters
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	dLat := rad(lat2 - lat1)
	dLon := rad(lon2 - lon1)
	lat1 = rad(lat1)
	lat2 = rad(lat2)

	// Apply haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadiusKm * c * 1000

	return math.Round(distance)
}
