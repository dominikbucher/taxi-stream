package taxisim

import (
	"math"
)

// Default earth radius.
var earthRadiusMetres float64 = 6371000

// Transforms a value in degrees to radians.
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// Computes the distance in meters from origin to destination using the Haversine formula.
func Distance(oLon float64, oLat float64, dLon float64, dLat float64) float64 {
	oLon = degreesToRadians(oLon)
	oLat = degreesToRadians(oLat)
	dLon = degreesToRadians(dLon)
	dLat = degreesToRadians(dLat)

	chgLon := oLon - dLon
	chgLat := oLat - dLat

	a := math.Pow(math.Sin(chgLat/2), 2) + math.Cos(oLat)*math.Cos(dLat)*math.Pow(math.Sin(chgLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return float64(earthRadiusMetres * c)
}
