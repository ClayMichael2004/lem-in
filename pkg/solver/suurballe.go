package solver

import (
	"errors"
	"math"
	"lem-in/pkg/antfarm"
)

func FindOptimalPaths(farm *antfarm.Farm) ([]antfarm.Path, error){
	// biuild split node flow graph with nodes having capacity of 1 and start and end being single nodes while the rest become r_in -> r_out
	graph:= newFlowGraph(farm)

	var bestPaths []antfarm.Path
	minTurns:= math.MaxInt32

	for{
		//find shortest path in residual graph using BFS
		augPath:= graph.findAugmentingPath(farm.Start.Name, farm.End.Name)
		if len(augPath)==0{
			break
		}

		//apply flow/ residual edges
		graph.applyPath(augPath)

		//extract current set of node-disjoint paths
		currentPaths:= graph.extractDisjointPaths(farm.Start.Name, farm.End.Name)
		if len(currentPaths)==0{
			continue
		}

		//calculate turns needed for N ants on current path set
		turns:= CalculateTurns(currentPaths, farm.Ants)
		if turns<minTurns{
			minTurns=turns
			bestPaths=currentPaths
		}else{
			break
		}
	}
	if len(bestPaths) == 0 {
		return nil, errors.New("no paths found between start and end")
	}

	return bestPaths, nil
}


//calculate the total turns required tpo move n ants across a set of paths
func CalculateTurns(paths []antfarm.Path, totalAnts int)int{
	if len(paths)==0{
		return math.MaxInt32
	}

	//calculating total capacity given path lengths
	maxPathLen:=0
	for _, p:= range paths{
		if len(p)>maxPathLen{
			maxPathLen=len(p)
		}
	}

	sumDiffs:=0
	for _,p:= range paths{
		sumDiffs+=(maxPathLen-len(p))
	}

	remainingAnts:= totalAnts-sumDiffs
	if remainingAnts<=0{
		return maxPathLen
	}

	turns:= maxPathLen+  int(math.Ceil(float64(remainingAnts)/float64(len(paths))))-1
	return turns
}

type flowGraph struct{
	capacity  map[string]map[string]int
	flow  	map[string]map[string]int
	adj 	 map[string][]string
}

func newFlowGraph(farm *antfarm.Farm) *flowGraph{
	fg:= &flowGraph{
		capacity: make(map[string]map[string]int),
		flow: make(map[string]map[string]int),
		adj: make(map[string][]string),
	}

	//split intermediate rooms
	for name, room:= range farm.Rooms{
		if room.IsStart ||  room.IsEnd {
			continue
		}
		inNode:= name +"_in"
		outNode:= name + "_out"
		fg.addEdge(inNode, outNode, 1)
	}

	//add the links to the rooms
	for u, neighbors:= range farm.Links{
		uRoom:= farm.Rooms[u]
		uFrom:=u
		if !uRoom.IsStart && !uRoom.IsEnd{
			uFrom=u + "_out"
		}
		for _, v:= range neighbors{
			vRoom:= farm.Rooms[v]
			vTo:=v
			if !vRoom.IsStart && !vRoom.IsEnd{
				vTo=v+"_in"
			}
			fg.addEdge(uFrom, vTo, 1)
		}
	}
	return fg
}

func (fg *flowGraph) addEdge(u, v string, cap int){
	if fg.capacity[u]==nil{
		fg.capacity[u]=make(map[string]int)
		fg.flow[u]=make(map[string]int)
	}
	if fg.capacity[v]==nil{
		fg.capacity[v]=make(map[string]int)
		fg.flow[v]=make(map[string]int)
	}
	fg.capacity[u][v]=cap
	fg.adj[u]=append(fg.adj[u], v)
	fg.adj[v]=append(fg.adj[v], u)
}

func (fg *flowGraph) findAugmentingPath(start, end string) []string{
	parent:= make(map[string]string)
	visited:=make(map[string]bool)
	queue:= []string{start}
	visited[start]=true

	for len(queue)>0{
		curr:= queue[0]
		queue = queue[1:]

		if curr==end{
			break
		}

		for _, neighbor:= range fg.adj[curr]{
			residual:= fg.capacity[curr][neighbor] - fg.flow[curr][neighbor]
			if residual > 0 && !visited[neighbor]{
				visited[neighbor]=true
				parent[neighbor]=curr
				queue=append(queue, neighbor)
			}
		}
	}

	if !visited[end]{
		return nil
	}

	//Reconstruct path
	var path []string
	curr:=end
	for curr!=start{
		prev:= parent[curr]
		path=append([]string{curr}, path...)
		curr=prev
	}
	return append([]string{start}, path...)
}

func (fg *flowGraph) applyPath(path []string){
	for i:=0; i<len(path)-1; i++{
		u, v:= path[i], path[i+1]
		fg.flow[u][v]+=1
		fg.flow[v][u]-=1
	}
}

func (fg *flowGraph) extractDisjointPaths(start, end string) []antfarm.Path{
	var paths []antfarm.Path

	for{
		path := fg.findFlowPath(start, end)
		if len(path)==0{
			break
		}
		paths=append(paths, path)
	}

	//restore flow state for further iterations
	for u:=range fg.flow{
		for v, f:= range fg.flow[u]{
			if f==-1{
				fg.flow[u][v]=1
			}
		}
	}
	return paths
}

func (fg *flowGraph) findFlowPath(start, end string) antfarm.Path {
	curr := start
	var rawPath []string

	for curr != end {
		next := ""

		// STRATEGIC FIX: If we are at an "_in" node, we MUST jump straight to its "_out" node
		if len(curr) > 3 && curr[len(curr)-3:] == "_in" {
			outNode := curr[:len(curr)-3] + "_out"
			if fg.flow[curr][outNode] == 1 {
				next = outNode
			}
		}

		// If no direct structural jump was forced, search for the next active flow edge
		if next == "" {
			for neighbor, f := range fg.flow[curr] {
				if f == 1 {
					next = neighbor
					break
				}
			}
		}

		if next == "" {
			return nil
		}

		fg.flow[curr][next] = -1
		curr = next
		rawPath = append(rawPath, curr)
	}

	// Clean split-node suffixes (_in / _out)
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
	return cleanPath
}
