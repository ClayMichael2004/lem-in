package parser

import (
	"strings"
	"testing"
)

func TestParseValidInput(t *testing.T) {
	input := `3
##start
start 0 0
##end
end 1 1
room1 0 1
start-room1
room1-end`

	farm, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("expected valid parse, got error: %v", err)
	}

	if farm.Ants != 3 {
		t.Errorf("expected 3 ants, got %d", farm.Ants)
	}
	if farm.Start.Name != "start" || farm.End.Name != "end" {
		t.Errorf("start or end room mismatch")
	}
}

func TestParseInvalidInputs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Invalid Ants", "0\n##start\n0 0 0\n##end\n1 1 1\n0-1"},
		{"Room starts with L", "2\n##start\nL1 0 0\n##end\n1 1 1\nL1-1"},
		{"Duplicate Room Names", "2\n##start\nA 0 0\nA 1 1\n##end\nB 2 2\nA-B"},
		{"Duplicate Room Coordinates", "2\n##start\nA 0 0\nB 0 0\n##end\nC 2 2\nA-B\nB-C"},
		{"Missing End", "2\n##start\nA 0 0\nB 1 1\nA-B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseReader(strings.NewReader(tt.input))
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.name)
			}
		})
	}
}