# lem-in

A modular Go-based pathfinding and simulation engine implementing advanced graph algorithms (Suurballe's algorithm) for ant colony optimization, featuring a dedicated visualizer component.

---

## Tech Stack Summary

| Category | Technology |
| :--- | :--- |
| **Language** | Go (`go.mod` present) |
| **Architecture** | Modular Monolith |
| **Core Packages** | `antfarm`, `parser`, `solver`, `visualizer` |

---

## Project Description

`lem-in` is a high-performance Go application designed to simulate and optimize the movement of ants through a colony graph. The system reads structured map layouts defining rooms, connections, and ant counts, parses and validates the topology, computes optimal paths using advanced graph algorithms (such as Suurballe's edge-disjoint path algorithm), and provides both a core simulation engine and a visualizer interface.

---

## Key Features

- **Robust Parser (`pkg/parser`)**: Ingests and validates colony configurations, rooms, tunnels, and ant populations from input streams and test data.
- **Advanced Graph Solver (`pkg/solver`)**: Implements Suurballe's algorithm (`suurballe.go`) alongside simulation logic (`simulation.go`) to find optimal disjoint paths for maximum throughput.
- **Ant Colony Management (`pkg/antfarm`)**: Encapsulates the state and structural rules of the ant farm ecosystem.
- **Web Visualizer (`cmd/web-visualizer`, `web/`)**: Go backend API server rendering an interactive, colorblind-friendly canvas visualization.
- **Comprehensive Test Suite**: Includes dedicated unit and integration tests across all core packages alongside real-world and edge-case test datasets (`testdata/`).

---

## Project Structure

```text
lem-in/
├── cmd/
│   ├── lem-in/            # Core CLI application entrypoint
│   │   └── main.go
│   ├── visualizer/        # Terminal visualizer entrypoint
│   │   └── main.go
│   └── web-visualizer/    # Go HTTP web visualizer server
│       └── main.go
├── pkg/
│   ├── antfarm/
│   │   ├── farm.go
│   │   └── farm_test.go
│   ├── parser/
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── solver/
│   │   ├── simulation.go
│   │   ├── solver_test.go
│   │   └── suurballe.go
│   └── visualizer/
│       ├── visualizer.go
│       └── visualizer_test.go
├── web/                   # Web interface templates and static assets
│   ├── index.html
│   └── static/
│       ├── css/style.css
│       └── js/visualizer.js
├── testdata/              # Benchmark and bad input test datasets
├── documentation.md       # Comprehensive technical documentation
├── go.mod
├── lem-in.exe
└── visualizer.exe
```

---

## Environment Configuration

No specific environment variables are required to run or test this project. 

If a local configuration is needed, verify that your environment supports standard Go execution paths (`GOPATH`, `GOROOT`).

---

## Installation & Getting Started

### Prerequisites

- Go (version 1.18 or higher recommended)

---

## Auditor Quick Start (Copy & Paste)

### 1. Build Binaries

**On Linux / macOS:**
```bash
go build -o lem-in ./cmd/lem-in
go build -o visualizer ./cmd/visualizer
```

**On Windows (PowerShell / CMD):**
```powershell
go build -o lem-in.exe ./cmd/lem-in
go build -o visualizer.exe ./cmd/visualizer
```

---

### 2. Run `lem-in` Simulation

**Linux / macOS:**
```bash
./lem-in testdata/example00.txt
```

**Windows (PowerShell):**
```powershell
./lem-in testdata/example00.txt
```

**Universal (Direct via `go run` on any OS):**
```bash
go run ./cmd/lem-in testdata/example00.txt
```

---

### 3. Run `visualizer` (Terminal Visualizer)

**Linux / macOS:**
```bash
./lem-in testdata/example00.txt | ./visualizer
```

**Windows (PowerShell):**
```powershell
./lem-in testdata/example00.txt | ./visualizer
```

**Universal (Direct via `go run` on any OS):**
```bash
go run ./cmd/lem-in testdata/example00.txt | go run ./cmd/visualizer
```

---

### 4. Interactive Web Visualizer (Go Backend Powered)

Run via Go HTTP server:
```bash
go run ./cmd/web-visualizer
```
Open **`http://localhost:8080`** in your browser.

---

### 5. Run Unit Tests
```bash
go test -v ./pkg/...
```

---

## Usage Examples

### Standard Test Cases
```bash
./lem-in testdata/example00.txt
./lem-in testdata/example01.txt
./lem-in testdata/example02.txt
./lem-in testdata/example03.txt
./lem-in testdata/example04.txt
./lem-in testdata/example05.txt
./lem-in testdata/example06.txt
./lem-in testdata/example07.txt
```

### Error Testing (Invalid Formats)
```bash
./lem-in testdata/badexample00.txt
./lem-in testdata/badexample01.txt
```

---

## Testing & Quality Assurance

To execute the test suite across all packages (`antfarm`, `parser`, `solver`, and `visualizer`), run:

```bash
go test -v ./pkg/...
```

To run tests with coverage reporting:

```bash
go test -cover ./pkg/...
```

---

## License

This project is distributed under standard licensing terms as defined in the repository.

## Authors

- Primary Maintainers & Contributors

---

## How the program works (high-level)

1. Input is provided as a `lem-in`-formatted map file (see `testdata/` examples). The main CLI (`cmd/lem-in/main.go`) reads a file path and passes the input to the parser.

2. Parsing: `pkg/parser` (`ParseFile` / `ParseReader`) validates the input and builds an `*antfarm.Farm` containing:
   - `Ants`: number of ants
   - `Rooms`: map of room names to `Room` (with coordinates and start/end flags)
   - `Links`: undirected adjacency between rooms
   - `RawInput`: the original input text

3. Path finding: `pkg/solver/suurballe.go` (`FindOptimalPaths`) constructs a flow graph from the farm and finds node-disjoint paths from the `##start` room to the `##end` room. It iteratively finds augmenting paths, extracts disjoint routes, and chooses the set that minimizes total turns for the given ant count.

4. Dispatch & Simulation: `pkg/solver/simulation.go` (`DispatchAndSimulate`) takes the chosen paths and the total ant count and computes turn-by-turn movements. It:
   - Sorts paths by length and pre-computes an optimal distribution of ants per path
   - Each turn, advances active ants along their paths and dispatches new ants onto available paths
   - Produces an ordered list of movement lines (e.g. `L1-roomA L2-roomB`) where each element represents one simulation turn

5. Output: `cmd/lem-in/main.go` prints the original `RawInput` and then each line returned by `DispatchAndSimulate`. Consumers (or the visualizer) can pipe this output for rendering.

6. Visualizer: `cmd/visualizer/main.go` calls `pkg/visualizer.RunVisualizer`, which reads the output from stdin and renders terminal frames showing ants moving between rooms.

7. Web Visualizer: `cmd/web-visualizer/main.go` hosts a HTTP web server serving static assets (`web/`) and providing REST endpoints (`/api/maps`, `/api/simulate`) powered directly by the Go parser and solver.

---

## Files of interest

- [cmd/lem-in/main.go](cmd/lem-in/main.go) — CLI entrypoint for parsing and running the simulation
- [cmd/visualizer/main.go](cmd/visualizer/main.go) — terminal visualizer CLI entrypoint
- [cmd/web-visualizer/main.go](cmd/web-visualizer/main.go) — web visualizer Go HTTP server entrypoint
- [pkg/antfarm/farm.go](pkg/antfarm/farm.go) — core domain types (`Farm`, `Room`, `Path`)
- [pkg/parser/parser.go](pkg/parser/parser.go) — input parsing and validation
- [pkg/solver/suurballe.go](pkg/solver/suurballe.go) — flow graph and path-finding (`FindOptimalPaths`)
- [pkg/solver/simulation.go](pkg/solver/simulation.go) — ant dispatch and simulation (`DispatchAndSimulate`)
- [pkg/visualizer/visualizer.go](pkg/visualizer/visualizer.go) — terminal visualizer rendering (`RunVisualizer`)
- [documentation.md](documentation.md) — comprehensive technical documentation and walkthroughs

