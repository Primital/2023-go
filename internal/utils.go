package internal

import (
	"math"

	"2023-go/types"
)

// Earth's radius in kilometers
const earthRadiusKm = 6371

// rad converts degrees to radians
func rad(deg float64) float64 {
	return deg * math.Pi / 180
}

// haversine calculates the distance between two geographic coordinates in meters
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
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

func DistanceBetweenCoordinates(lat1, lon1, lat2, lon2 float64) float64 {
	R := 6371e3
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	d := R * c

	return RoundFloatByN(d, 0)
}

func RoundFloatByN(x float64, n int) float64 {
	// If n i 2, and x is 1.2345, then this function should return 1.23
	// If n is 3, and x is 1.2345, then this function should return 1.235
	return math.Round(x*math.Pow10(n)) / math.Pow10(n)
}

func PrecalculateNeighborDistances(locations []*types.Location, data *types.GeneralGameData) {
	for _, loc := range locations {
		loc.NeighborDistances = make(map[string]float64)
		for _, other := range locations {
			if loc.Name == other.Name {
				continue
			}
			dist := Haversine(loc.Latitude, loc.Longitude, other.Latitude, other.Longitude)
			if dist < data.WillingnessToTravelInMeters {
				loc.NeighborDistances[other.Name] = dist
			}
		}
	}
}
