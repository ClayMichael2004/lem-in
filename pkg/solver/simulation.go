package solver

import (
	"fmt"
	"strings"
	"sort"
	"lem-in/pkg/antfarm"
)

type Ant struct{
	ID int
	Path antfarm.Path
	StepIndex int
}

func DispatchAndSimulate(paths []antfarm.Path, totalAnts int) []string{
	//sort paths by length in ascending order
	sort.Slice(paths, func(i, j int)bool{
		return len(paths[i])< len(paths[j])
	})

	//determine how many ants go down each path
	pathAssignments:= make([]int, len(paths))
	for ant:= 0; ant<totalAnts; ant++{
		bestPathIdx:=0
		minCost:=len(paths[0])+pathAssignments[0]

		for i:=1;i<len(paths); i++{
			cost:=len(paths[i])+pathAssignments[i]
			if cost< minCost{
				minCost=cost
				bestPathIdx=i
			}
		}
		pathAssignments[bestPathIdx]++
	}

	//assign paths to ant objects
	ants:=make([]*Ant, totalAnts)
	antID:=1

	for pathIdx, count:=range paths{
		numAntsForPath:=pathAssignments[pathIdx]
		for k:=0; k< numAntsForPath; k++{
			ants[antID-1]=&Ant{
				ID: antID,
				Path: count,
				StepIndex: -1,
			}
			antID++
		}
	}

	var turnOutputs []string
	activeAnts:=[]*Ant{}
	nextAntToDeploy:=0

	for {
		var movesThisTurn []antfarm.Movement
		var remainingActive []*Ant

		//advance existing active ants
		for _, ant:= range activeAnts{
			ant.StepIndex++
			movesThisTurn= append(movesThisTurn, antfarm.Movement{
				AntID:    ant.ID,
				RoomName:  ant.Path[ant.StepIndex],
			})
			if ant.StepIndex< len(ant.Path)-1{
				remainingActive=append(remainingActive, ant)
			}
		}
		activeAnts=remainingActive

		//deploy new ants on eligible paths
		pathsUsedThisTurn:= make(map[int]bool)
		for nextAntToDeploy< totalAnts{
			ant:=ants[nextAntToDeploy]

			pathIdx:=-1
			for idx, p:=range paths{
				if len(p)==len(ant.Path) && &p[0]==&ant.Path[0]{
					pathIdx=idx
					break
				}
			}

			if pathIdx != -1 && pathsUsedThisTurn[pathIdx]{
				break
			}

			ant.StepIndex=0
			movesThisTurn=append(movesThisTurn, antfarm.Movement{
				AntID:   ant.ID,
				RoomName: ant.Path[0],
			})

			if pathIdx != -1{
				pathsUsedThisTurn[pathIdx]=true
			}

			if len(ant.Path)>1{
				activeAnts=append(activeAnts, ant)
			}
			nextAntToDeploy++
		}

		if len(movesThisTurn)==0{
			break
		}

		//format output line: L1-roomA L2-roomB
		var moveStrings[]string
		for _, m:= range movesThisTurn{
			moveStrings= append(moveStrings, fmt.Sprintf("L%d-%s", m.AntID, m.RoomName))
		}
		turnOutputs=append(turnOutputs, strings.Join(moveStrings, " "))
	}
	return turnOutputs
}