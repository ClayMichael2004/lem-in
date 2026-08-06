package visualizer

import (
	"strings"
	"testing"
)

func TestRunVisualizer_Success(t *testing.T) {
	// Valid lem-in output format containing rooms and ant movements
	input := `
##start
start 0 0
roomA 2 0
roomB 0 2
##end
end 2 2
start-roomA
roomA-end

L1-roomA L2-roomB
L1-end L2-end
`

	r := strings.NewReader(input)
	err := RunVisualizer(r)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunVisualizer_NoRoomsError(t *testing.T) {
	// Input with movements but missing valid room coordinates
	input := `
L1-roomA L2-roomB
`

	r := strings.NewReader(input)
	err := RunVisualizer(r)
	if err == nil {
		t.Fatal("expected error when no room coordinates are parsed, but got nil")
	}

	expectedSub := "no room coordinates"
	if !strings.Contains(err.Error(), expectedSub) {
		t.Errorf("expected error containing %q, got %q", expectedSub, err.Error())
	}
}

func TestRunVisualizer_MalformedInput(t *testing.T) {
	// Input with invalid integers for room coordinates
	input := `
roomA abc def
roomB 10 ghi
`

	r := strings.NewReader(input)
	err := RunVisualizer(r)
	if err == nil {
		t.Fatal("expected error due to invalid room coordinates, got nil")
	}
}

func TestRunVisualizer_ZeroTurnMoves(t *testing.T) {
	// Valid room coordinates, but no ant moves recorded
	input := `
start 0 0
end 5 5
`

	r := strings.NewReader(input)
	err := RunVisualizer(r)
	if err != nil {
		t.Fatalf("expected program to exit cleanly with 0 turns, got: %v", err)
	}
}

func TestRunVisualizer_SingleRoomGridScaling(t *testing.T) {
	// Edge case where minX == maxX and minY == maxY to ensure no division by zero in renderGrid
	input := `
roomSingle 5 5
L1-roomSingle
`

	r := strings.NewReader(input)
	err := RunVisualizer(r)
	if err != nil {
		t.Fatalf("expected program to handle single-room grid bounds without panic, got: %v", err)
	}
}