package solver

import (
	"errors"
	"math"

	"lem-in/pkg/antfarm"
)

// FindOptimalPaths finds the optimal set of node-disjoint paths minimizing total turns.
func FindOptimalPaths(farm *antfarm.Farm) ([]antfarm.Path, error) {
	graph := newFlowGraph(farm)

	var bestPaths []antfarm.Path
	minTurns := math.MaxInt32

	for {
		augPath := graph.findAugmentingPath(farm.Start.Name, farm.End.Name)
		if len(augPath) == 0 {
			break
		}

		graph.applyPath(augPath)

		currentPaths := graph.extractDisjointPaths(farm.Start.Name, farm.End.Name)
		if len(currentPaths) == 0 {
			continue
		}

		turns := CalculateTurns(currentPaths, farm.Ants)
		if turns < minTurns {
			minTurns = turns
			bestPaths = currentPaths
		}
	}

	if len(bestPaths) == 0 {
		return nil, errors.New("ERROR: invalid data format, no path found between ##start and ##end")
	}

	return bestPaths, nil
}

// CalculateTurns calculates total turns required to move totalAnts across paths.
func CalculateTurns(paths []antfarm.Path, totalAnts int) int {
	if len(paths) == 0 {
		return math.MaxInt32
	}

	maxPathLen := 0
	for _, p := range paths {
		if len(p) > maxPathLen {
			maxPathLen = len(p)
		}
	}

	sumDiffs := 0
	for _, p := range paths {
		sumDiffs += (maxPathLen - len(p))
	}

	remainingAnts := totalAnts - sumDiffs
	if remainingAnts <= 0 {
		return maxPathLen
	}

	return maxPathLen + int(math.Ceil(float64(remainingAnts)/float64(len(paths)))) - 1
}

type flowGraph struct {
	capacity map[string]map[string]int
	flow     map[string]map[string]int
	adj      map[string][]string
}

func newFlowGraph(farm *antfarm.Farm) *flowGraph {
	fg := &flowGraph{
		capacity: make(map[string]map[string]int),
		flow:     make(map[string]map[string]int),
		adj:      make(map[string][]string),
	}

	for name, room := range farm.Rooms {
		if room.IsStart || room.IsEnd {
			continue
		}
		fg.addEdge(name+"_in", name+"_out", 1)
	}

	for u, neighbors := range farm.Links {
		uRoom := farm.Rooms[u]
		uFrom := u
		if uRoom != nil && !uRoom.IsStart && !uRoom.IsEnd {
			uFrom = u + "_out"
		}

		for _, v := range neighbors {
			vRoom := farm.Rooms[v]
			vTo := v
			if vRoom != nil && !vRoom.IsStart && !vRoom.IsEnd {
				vTo = v + "_in"
			}
			fg.addEdge(uFrom, vTo, 1)
		}
	}

	return fg
}

func (fg *flowGraph) addEdge(u, v string, cap int) {
	if fg.capacity[u] == nil {
		fg.capacity[u] = make(map[string]int)
		fg.flow[u] = make(map[string]int)
	}
	if fg.capacity[v] == nil {
		fg.capacity[v] = make(map[string]int)
		fg.flow[v] = make(map[string]int)
	}
	fg.capacity[u][v] = cap
	fg.adj[u] = append(fg.adj[u], v)
	fg.adj[v] = append(fg.adj[v], u)
}

func (fg *flowGraph) findAugmentingPath(start, end string) []string {
	parent := make(map[string]string)
	visited := make(map[string]bool)
	queue := []string{start}
	visited[start] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == end {
			break
		}

		for _, neighbor := range fg.adj[curr] {
			residual := fg.capacity[curr][neighbor] - fg.flow[curr][neighbor]
			if residual > 0 && !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = curr
				queue = append(queue, neighbor)
			}
		}
	}

	if !visited[end] {
		return nil
	}

	var path []string
	curr := end
	for curr != start {
		prev := parent[curr]
		path = append([]string{curr}, path...)
		curr = prev
	}
	return append([]string{start}, path...)
}

func (fg *flowGraph) applyPath(path []string) {
	for i := 0; i < len(path)-1; i++ {
		u, v := path[i], path[i+1]
		fg.flow[u][v] += 1
		fg.flow[v][u] -= 1
	}
}

func (fg *flowGraph) extractDisjointPaths(start, end string) []antfarm.Path {
	var paths []antfarm.Path

	flowCopy := make(map[string]map[string]int)
	for u, neighbors := range fg.flow {
		flowCopy[u] = make(map[string]int)
		for v, f := range neighbors {
			flowCopy[u][v] = f
		}
	}

	for {
		parent := make(map[string]string)
		visited := make(map[string]bool)
		queue := []string{start}
		visited[start] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			if curr == end {
				break
			}

			for neighbor, f := range flowCopy[curr] {
				if f > 0 && !visited[neighbor] {
					visited[neighbor] = true
					parent[neighbor] = curr
					queue = append(queue, neighbor)
				}
			}
		}

		if !visited[end] {
			break
		}

		var rawPath []string
		curr := end
		for curr != start {
			prev := parent[curr]
			flowCopy[prev][curr]--
			rawPath = append([]string{curr}, rawPath...)
			curr = prev
		}

		var cleanPath antfarm.Path
		for _, node := range rawPath {
			cleanNode := node
			if len(node) > 3 && node[len(node)-3:] == "_in" {
				cleanNode = node[:len(node)-3]
			} else if len(node) > 4 && node[len(node)-4:] == "_out" {
				cleanNode = node[:len(node)-4]
			}

			if len(cleanPath) == 0 || cleanPath[len(cleanPath)-1] != cleanNode {
				cleanPath = append(cleanPath, cleanNode)
			}
		}

		paths = append(paths, cleanPath)
	}

	return paths
}