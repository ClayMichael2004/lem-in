// pkg/solver/solver_test.go
package solver

import (
	"testing"
	"lem-in/pkg/parser"
)

func TestExample00TurnLimit(t *testing.T) {
	farm, _, err := parser.ParseFile("../../testdata/example00.txt")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	paths, err := FindOptimalPaths(farm)
	if err != nil {
		t.Fatalf("Failed to solve: %v", err)
	}
	turns := CalculateTurnCount(paths, farm.Ants)
	if turns > 6 {
		t.Errorf("Expected <= 6 turns, got %d", turns)
	}
}