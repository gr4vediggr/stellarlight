package utils_test

import (
	"testing"

	"github.com/gr4vediggr/stellarlight/internal/utils"
)

type MockChoice struct {
	name   string
	weight float64
}

func (m *MockChoice) GetChoiceWeight() float64 {
	return m.weight
}

func TestWeightedRandomChance(t *testing.T) {

	choices := []*MockChoice{
		{name: "A", weight: 0.5},
		{name: "B", weight: 1.5},
		{name: "C", weight: 3.0},
	}

	results := make(map[string]int)
	for i := 0; i < 1000; i++ {
		choice := utils.WeightedRandomChoice(choices)
		results[choice.name]++
	}

	if results["A"] > results["B"] || results["B"] > results["C"] {
		t.Errorf("Unexpected distribution of choices: %v", results)
	}

	if results["C"]/results["A"] < 5 {
		t.Errorf("Choice C should be chosen more frequently than A, got: %v", results)
	}
}
