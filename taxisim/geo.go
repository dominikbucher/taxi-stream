package taxisim

import (
	"math"
	"errors"
)

// Default earth radius.
var earthRadiusMetres float64 = 6371000

// Transforms a value in degrees to radians.
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// Computes the distance in meters from origin to destination using the Haversine formula.
func HaversineDistance(oLon float64, oLat float64, dLon float64, dLat float64) float64 {
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

// Computes the Euclidean distance between points.
func Distance(lon1 float64, lat1 float64, lon2 float64, lat2 float64) float64 {
	return math.Sqrt((lon1-lon2)*(lon1-lon2) + (lat1-lat2)*(lat1-lat2))
}

// Computes the (Euclidean) length of a polyline.
func PolylineLength(coords [][]float64) float64 {
	var totDist float64 = 0
	for i := 0; i < len(coords)-1; i++ {
		c1 := coords[i]
		c2 := coords[i+1]
		d := Distance(c1[1], c1[0], c2[1], c2[0])
		totDist += d
	}
	return totDist
}

// Computes a coordinate along a polyline, namely after "dist" into the polyline.
func AlongPolyline(dist float64, coords [][]float64) (float64, float64) {
	var totDist float64 = 0
	for i := 0; i < len(coords)-1; i++ {
		c1 := coords[i]
		c2 := coords[i+1]
		d := Distance(c1[1], c1[0], c2[1], c2[0])
		if totDist+d > dist {
			perc := d / (dist - totDist)
			lon := c1[1] + (c2[1]-c1[1])*perc
			lat := c1[0] + (c2[0]-c1[0])*perc
			return lon, lat
		} else {
			totDist += d
		}
	}
	panic(errors.New("dist longer than polyline"))
}
