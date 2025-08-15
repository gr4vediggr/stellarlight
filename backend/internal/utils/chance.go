package utils

import "math/rand/v2"

type WeightedChance interface {
	GetChoiceWeight() float64
}

func WeightedRandomChoice[T WeightedChance](choices []T) T {
	if len(choices) == 0 {
		var zero T
		return zero
	}

	totalWeight := 0.0
	for _, choice := range choices {
		totalWeight += choice.GetChoiceWeight()
	}

	randomValue := rand.Float64() * totalWeight
	for _, choice := range choices {
		randomValue -= choice.GetChoiceWeight()
		if randomValue <= 0 {
			return choice
		}
	}

	return choices[len(choices)-1] // Fallback in case of rounding errors
}

func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	return int(rand.Int32N(int32(max-min))) + min
}

func RandomFloat(min, max float64) float64 {
	if min >= max {
		return min
	}
	return rand.Float64()*(max-min) + min
}
