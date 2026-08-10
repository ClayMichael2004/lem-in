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
- **Visualizer (`cmd/visualizer`, `pkg/visualizer`)**: Renders graphical or structural representations of the simulation steps.
- **Comprehensive Test Suite**: Includes dedicated unit and integration tests across all core packages alongside real-world and edge-case test datasets (`testdata/`).

---

## Project Structure

```text
lem-in/
├── cmd/
│   ├── lem-in/
│   │   └── main.go
│   └── visualizer/
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
├── testdata/
│   ├── badexample00.txt
│   ├── badexample01.txt
│   ├── example00.txt
│   ├── example01.txt
│   ├── example02.txt
│   ├── example03.txt
│   ├── example04.txt
│   ├── example05.txt
│   ├── example06.txt
│   └── example07.txt
├── go.mod
├── lem-in (binary)
└── visualizer (binary)
```

---

## Environment Configuration

No specific environment variables are required to run or test this project. 

If a local configuration is needed, verify that your environment supports standard Go execution paths (`GOPATH`, `GOROOT`).

---

## Installation & Getting Started

### Prerequisites

- Go (version 1.18 or higher recommended)

### Setup Instructions

1. Navigate into the project directory:
   ```bash
   cd lem-in
   ```

2. Build the main application and visualizer binaries:
   ```bash
   go build -o lem-in ./cmd/lem-in
   go build -o visualizer ./cmd/visualizer
   ```

3. Run the main simulation engine against an example map:
   ```bash
   ./lem-in < testdata/example00.txt
   ```

---

## Available Scripts & Commands

Because no explicit script runner or package manager wrapper is defined outside of standard Go toolchains, use standard Go commands:

- **Build Core Engine**: `go build -o lem-in ./cmd/lem-in`
- **Build Visualizer**: `go build -o visualizer ./cmd/visualizer`
- **Run Engine**: `./lem-in < [path-to-map-file]`

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