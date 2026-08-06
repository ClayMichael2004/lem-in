package visualizer

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type RoomCoord struct {
	Name string
	X    int
	Y    int
}

type StepMove struct {
	AntID    int
	RoomName string
}

// RunVisualizer reads piped stdout from standard input and renders the simulation turn-by-turn.
func RunVisualizer(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	rooms := make(map[string]RoomCoord)
	var turns [][]StepMove

	minX, minY := math.MaxInt32, math.MaxInt32
	maxX, maxY := math.MinInt32, math.MinInt32

	readingMoves := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Detect start of movements lines (e.g., L1-roomA L2-roomB)
		if strings.HasPrefix(line, "L") && strings.Contains(line, "-") {
			readingMoves = true
		}

		if !readingMoves {
			// Ignore comments and commands
			if strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Fields(line)
			// Parse rooms format: "name coord_x coord_y"
			if len(parts) == 3 {
				x, errX := strconv.Atoi(parts[1])
				y, errY := strconv.Atoi(parts[2])
				if errX == nil && errY == nil {
					rooms[parts[0]] = RoomCoord{Name: parts[0], X: x, Y: y}
					if x < minX {
						minX = x
					}
					if x > maxX {
						maxX = x
					}
					if y < minY {
						minY = y
					}
					if y > maxY {
						maxY = y
					}
				}
			}
		} else {
			// Parse turn movements: "L1-3 L2-2"
			tokens := strings.Fields(line)
			var turn []StepMove
			for _, tok := range tokens {
				if strings.HasPrefix(tok, "L") {
					parts := strings.Split(tok[1:], "-")
					if len(parts) == 2 {
						antID, err := strconv.Atoi(parts[0])
						if err == nil {
							turn = append(turn, StepMove{
								AntID:    antID,
								RoomName: parts[1],
							})
						}
					}
				}
			}
			if len(turn) > 0 {
				turns = append(turns, turn)
			}
		}
	}

	if len(rooms) == 0 {
		return fmt.Errorf("ERROR: visualizer received no room coordinates")
	}

	// Render the simulation frame-by-frame
	antPositions := make(map[int]string)

	clearTerminal()
	fmt.Println("=== LEM-IN ANT FARM VISUALIZER ===")
	time.Sleep(1 * time.Second)

	for turnIdx, turn := range turns {
		// Update ant positions
		for _, move := range turn {
			antPositions[move.AntID] = move.RoomName
		}

		clearTerminal()
		fmt.Printf("--- TURN %d / %d ---\n\n", turnIdx+1, len(turns))
		renderGrid(rooms, antPositions, minX, maxX, minY, maxY)
		time.Sleep(400 * time.Millisecond)
	}

	fmt.Println("\nAll ants have successfully crossed the colony!")
	return nil
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func renderGrid(rooms map[string]RoomCoord, antPositions map[int]string, minX, maxX, minY, maxY int) {
	// Normalize scale for terminal display
	width := maxX - minX + 1
	height := maxY - minY + 1

	if width > 40 {
		width = 40
	}
	if height > 20 {
		height = 20
	}

	// Build grid
	grid := make([][]string, height+1)
	for i := range grid {
		grid[i] = make([]string, width+1)
		for j := range grid[i] {
			grid[i][j] = " . "
		}
	}

	// Invert room positions onto grid
	for roomName, coord := range rooms {
		gx := 0
		if maxX > minX {
			gx = (coord.X - minX) * width / (maxX - minX)
		}
		gy := 0
		if maxY > minY {
			gy = (coord.Y - minY) * height / (maxY - minY)
		}

		if gx >= 0 && gx <= width && gy >= 0 && gy <= height {
			grid[gy][gx] = fmt.Sprintf("[%s]", roomName[:min(len(roomName), 2)])
		}
	}

	// Display ants currently in rooms
	for antID, roomName := range antPositions {
		if coord, ok := rooms[roomName]; ok {
			gx := 0
			if maxX > minX {
				gx = (coord.X - minX) * width / (maxX - minX)
			}
			gy := 0
			if maxY > minY {
				gy = (coord.Y - minY) * height / (maxY - minY)
			}

			if gx >= 0 && gx <= width && gy >= 0 && gy <= height {
				grid[gy][gx] = fmt.Sprintf("A%d", antID%10)
			}
		}
	}

	for _, row := range grid {
		for _, cell := range row {
			fmt.Print(cell)
		}
		fmt.Println()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}