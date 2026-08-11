package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"lem-in/pkg/parser"
	"lem-in/pkg/solver"
)

type RoomData struct {
	Name    string `json:"name"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	IsStart bool   `json:"isStart"`
	IsEnd   bool   `json:"isEnd"`
}

type LinkData struct {
	U string `json:"u"`
	V string `json:"v"`
}

type MoveData struct {
	AntID int    `json:"antID"`
	Room  string `json:"room"`
}

type SimulationResponse struct {
	TotalAnts  int                 `json:"totalAnts"`
	StartRoom  string              `json:"startRoom"`
	EndRoom    string              `json:"endRoom"`
	Rooms      map[string]RoomData `json:"rooms"`
	Links      []LinkData          `json:"links"`
	Turns      [][]MoveData        `json:"turns"`
	TurnStates []map[int]string    `json:"turnStates"`
	RawInput   string              `json:"rawInput"`
}

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	webDir := filepath.Join(dir, "web")
	staticDir := filepath.Join(webDir, "static")

	// 1. Serve static files (CSS & JS)
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 2. Serve Index HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
	})

	// 3. API endpoint: List available maps in testdata/
	http.HandleFunc("/api/maps", func(w http.ResponseWriter, r *http.Request) {
		testdataDir := filepath.Join(dir, "testdata")
		entries, err := os.ReadDir(testdataDir)
		if err != nil {
			http.Error(w, "Unable to read testdata directory", http.StatusInternalServerError)
			return
		}

		var maps []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") && !strings.HasPrefix(entry.Name(), "bad") {
				maps = append(maps, entry.Name())
			}
		}
		sort.Strings(maps)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(maps)
	})

	// 4. API endpoint: Run Go parser & solver and return JSON simulation
	http.HandleFunc("/api/simulate", func(w http.ResponseWriter, r *http.Request) {
		mapName := r.URL.Query().Get("map")
		if mapName == "" {
			mapName = "example00.txt"
		}

		// Prevent path traversal
		cleanName := filepath.Base(mapName)
		filePath := filepath.Join(dir, "testdata", cleanName)

		farm, err := parser.ParseFile(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		paths, err := solver.FindOptimalPaths(farm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		movesOutput := solver.DispatchAndSimulate(paths, farm.Ants)

		// Convert Go antfarm structures to web JSON format
		roomsMap := make(map[string]RoomData)
		for name, room := range farm.Rooms {
			roomsMap[name] = RoomData{
				Name:    room.Name,
				X:       room.X,
				Y:       room.Y,
				IsStart: room.IsStart,
				IsEnd:   room.IsEnd,
			}
		}

		var linksList []LinkData
		seenLink := make(map[string]bool)
		for u, neighbors := range farm.Links {
			for _, v := range neighbors {
				key1 := u + "-" + v
				key2 := v + "-" + u
				if !seenLink[key1] && !seenLink[key2] {
					linksList = append(linksList, LinkData{U: u, V: v})
					seenLink[key1] = true
					seenLink[key2] = true
				}
			}
		}

		var parsedTurns [][]MoveData
		for _, line := range movesOutput {
			var turnMoves []MoveData
			tokens := strings.Fields(line)
			for _, tok := range tokens {
				if strings.HasPrefix(tok, "L") {
					dashIdx := strings.Index(tok, "-")
					if dashIdx > 1 {
						antID, err := strconv.Atoi(tok[1:dashIdx])
						roomName := tok[dashIdx+1:]
						if err == nil && roomName != "" {
							turnMoves = append(turnMoves, MoveData{
								AntID: antID,
								Room:  roomName,
							})
						}
					}
				}
			}
			if len(turnMoves) > 0 {
				parsedTurns = append(parsedTurns, turnMoves)
			}
		}

		// Compute turn-by-turn positions
		var turnStates []map[int]string
		currentPos := make(map[int]string)
		turnStates = append(turnStates, copyMap(currentPos))

		for _, turn := range parsedTurns {
			for _, move := range turn {
				currentPos[move.AntID] = move.Room
			}
			turnStates = append(turnStates, copyMap(currentPos))
		}

		resp := SimulationResponse{
			TotalAnts:  farm.Ants,
			StartRoom:  farm.Start.Name,
			EndRoom:    farm.End.Name,
			Rooms:      roomsMap,
			Links:      linksList,
			Turns:      parsedTurns,
			TurnStates: turnStates,
			RawInput:   farm.RawInput,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Printf("🚀 Lem-in Web Visualizer powered by Go backend is live at: http://localhost:%s\n", port)
	fmt.Println("Press Ctrl+C to stop.")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func copyMap(src map[int]string) map[int]string {
	dst := make(map[int]string)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
