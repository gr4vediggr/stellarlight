package gen

import "github.com/google/uuid"

func GenerateStarName() string {
	// This function generates a random star name.
	// For simplicity, we will return a static name here.
	// In a real application, you might want to use a more complex algorithm or a library.
	return "Star-" + uuid.New().String()[:8]
}

func GeneratePlanetName() string {
	// This function generates a random planet name.
	// For simplicity, we will return a static name here.
	// In a real application, you might want to use a more complex algorithm or a library.
	return "Planet-" + uuid.New().String()[:8]
}

func GenerateStarSystemName() string {
	// This function generates a random star system name.
	// For simplicity, we will return a static name here.
	// In a real application, you might want to use a more complex algorithm or a library.
	return "StarSystem-" + uuid.New().String()[:8]
}
