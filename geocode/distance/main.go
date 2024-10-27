package main

import (
	"fmt"
	"math"
)

var (
	// Example coordinates (latitude and longitude)
	lat1, lon1 = 40.7128, -74.0060  // New York City
	lat2, lon2 = 34.0522, -118.2437 // Los Angeles

	// Example coordinates for Seoul and Busan
	latSeoul, lonSeoul = 37.5665, 126.9780
	latBusan, lonBusan = 35.1796, 129.0756
)

func main() {

	distance := haversine(latSeoul, lonSeoul, latBusan, lonBusan)
	fmt.Printf("Distance: %.2f meters\n", distance)

	distance, err := vincenty(latSeoul, lonSeoul, latBusan, lonBusan)
	if err != nil {
		fmt.Println("Error calculating distance:", err)
		return
	}
	fmt.Printf("Distance between Seoul and Busan: %.2f meters\n", distance)
}

const (
	// earthRadius = 6371000.0 // Earth's radius in meters
	// earthRadius = 6378140.0 // Earth's radius in meters
	earthRadius = 6378137.0 // Earth's radius in meters
)

// haversine function calculates the distance between two geo-coordinates in meters.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	// const earthRadius = 6371000 // Earth's radius in meters

	// Convert latitude and longitude from degrees to radians.
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Calculate differences
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Apply the Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in meters
	return earthRadius * c
}

// vincenty calculates the distance between two latitude/longitude points using Vincenty's formula
func vincenty(lat1, lon1, lat2, lon2 float64) (float64, error) {
	// const a = 6378137.0                // Semi-major axis of the Earth (in meters)
	const f = 1 / 298.257223563        // Flattening of the Earth
	const b = (1 - f) * earthRadius    // Semi-minor axis of the Earth (in meters)
	const maxIterations = 1000         // Maximum number of iterations
	const convergenceThreshold = 1e-12 // Convergence threshold

	// Convert latitude and longitude from degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Calculate U1 and U2
	U1 := math.Atan((1 - f) * math.Tan(lat1Rad))
	U2 := math.Atan((1 - f) * math.Tan(lat2Rad))

	// Calculate the difference in longitude
	L := lon2Rad - lon1Rad

	var lambda = L
	var sinSigma, cosSigma, sigma, sinAlpha, cos2Alpha, cos2SigmaM, C float64

	for i := 0; i < maxIterations; i++ {
		// Calculate trigonometric terms
		sinLambda := math.Sin(lambda)
		cosLambda := math.Cos(lambda)
		sinSigma = math.Sqrt(math.Pow(math.Cos(U2)*sinLambda, 2) +
			math.Pow(math.Cos(U1)*math.Sin(U2)-math.Sin(U1)*math.Cos(U2)*cosLambda, 2))
		cosSigma = math.Sin(U1)*math.Sin(U2) + math.Cos(U1)*math.Cos(U2)*cosLambda
		sigma = math.Atan2(sinSigma, cosSigma)

		// Calculate the angle alpha
		sinAlpha = math.Cos(U1) * math.Cos(U2) * sinLambda / sinSigma
		cos2Alpha = 1 - sinAlpha*sinAlpha
		cos2SigmaM = cosSigma - 2*math.Sin(U1)*math.Sin(U2)/cos2Alpha

		// Handle the case where the line is equatorial
		if math.IsNaN(cos2SigmaM) {
			cos2SigmaM = 0 // Equatorial line: cos2Alpha=0
		}

		// Calculate C
		C = f / 16 * cos2Alpha * (4 + f*(4-3*cos2Alpha))

		// Update lambda
		lambdaPrev := lambda
		lambda = L + (1-C)*f*sinAlpha*(sigma+C*sinSigma*(cos2SigmaM+C*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)))

		// Check for convergence
		if math.Abs(lambda-lambdaPrev) < convergenceThreshold {
			break
		}
	}

	// Calculate the distance
	u2 := cos2Alpha * (earthRadius*earthRadius - b*b) / (b * b)
	A := 1 + u2/16384*(4096+u2*(-768+u2*(320-175*u2)))
	B := u2 / 1024 * (256 + u2*(-128+u2*(74-47*u2)))
	deltaSigma := B * sinSigma * (cos2SigmaM + B/4*(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)-
		B/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)))

	// Distance in meters
	distance := b * A * (sigma - deltaSigma)

	return distance, nil
}
