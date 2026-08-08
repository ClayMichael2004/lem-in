package solver

import (
	"fmt"
	"sort"
	"strings"

	"lem-in/pkg/antfarm"
)

type AntState struct {
	ID        int
	PathIdx   int
	StepIndex int
}

// DispatchAndSimulate allocates ants dynamically across paths at each turn.
func DispatchAndSimulate(paths []antfarm.Path, totalAnts int) []string {
	// 1. Sort paths by length ascending
	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i]) < len(paths[j])
	})

	// 2. Pre-calculate exact ant distribution per path to optimize turn count
	antsPerPath := make([]int, len(paths))
	for ant := 0; ant < totalAnts; ant++ {
		bestPath := 0
		minCost := len(paths[0]) + antsPerPath[0]
		for i := 1; i < len(paths); i++ {
			cost := len(paths[i]) + antsPerPath[i]
			if cost < minCost {
				minCost = cost
				bestPath = i
			}
		}
		antsPerPath[bestPath]++
	}

	var movesOutput []string
	var activeAnts []*AntState
	
	nextAntID := 1
	antsRemainingInPath := make([]int, len(paths))
	copy(antsRemainingInPath, antsPerPath)

	for {
		var movesThisTurn []string
		var nextActive []*AntState

		// A. Move active ants forward along their paths
		for _, ant := range activeAnts {
			ant.StepIndex++
			targetRoom := paths[ant.PathIdx][ant.StepIndex]
			movesThisTurn = append(movesThisTurn, fmt.Sprintf("L%d-%s", ant.ID, targetRoom))

			if ant.StepIndex < len(paths[ant.PathIdx])-1 {
				nextActive = append(nextActive, ant)
			}
		}
		activeAnts = nextActive

		// B. Dispatch 1 new ant onto EACH path that still has assigned ants
		for pathIdx := 0; pathIdx < len(paths); pathIdx++ {
			if antsRemainingInPath[pathIdx] > 0 && nextAntID <= totalAnts {
				ant := &AntState{
					ID:        nextAntID,
					PathIdx:   pathIdx,
					StepIndex: 0,
				}
				nextAntID++
				antsRemainingInPath[pathIdx]--

				targetRoom := paths[pathIdx][0]
				movesThisTurn = append(movesThisTurn, fmt.Sprintf("L%d-%s", ant.ID, targetRoom))

				if len(paths[pathIdx]) > 1 {
					activeAnts = append(activeAnts, ant)
				}
			}
		}

		if len(movesThisTurn) == 0 {
			break
		}

		movesOutput = append(movesOutput, strings.Join(movesThisTurn, " "))
	}

	return movesOutput
}