package solver

import (
	"strings"
	"testing"

	"lem-in/pkg/parser"
)

func TestTurnLimits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxTurns int
	}{
		{
			name: "example00",
			input: `4
##start
0 0 3
2 2 5
3 4 0
##end
1 8 3
0-2
2-3
3-1`,
			maxTurns: 6,
		},
		{
			name: "example03",
			input: `4
4 5 4
##start
0 1 4
1 3 6
##end
5 6 4
2 3 4
3 3 1
0-1
2-4
1-4
0-2
4-5
3-0
4-3`,
			maxTurns: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			farm, err := parser.ParseReader(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			paths, err := FindOptimalPaths(farm)
			if err != nil {
				t.Fatalf("unexpected solver error: %v", err)
			}

			turns := DispatchAndSimulate(paths, farm.Ants)
			if len(turns) > tt.maxTurns {
				t.Errorf("expected max %d turns, got %d turns", tt.maxTurns, len(turns))
			}
		})
	}
}