# Comprehensive Technical Documentation: `lem-in`

---

## Table of Contents

1. [Executive Overview & Objectives](#1-executive-overview--objectives)
2. [Domain Rules & Constraints](#2-domain-rules--constraints)
3. [Graph Theory & Algorithmic Concepts](#3-graph-theory--algorithmic-concepts)
   - [Flow Networks & Node-Disjoint Paths](#flow-networks--node-disjoint-paths)
   - [Node Splitting Technique](#node-splitting-technique)
   - [Edmonds-Karp / Suurballe's Algorithm](#edmonds-karp--suurballes-algorithm)
   - [Turn Calculation & Ant Allocation Math](#turn-calculation--ant-allocation-math)
4. [Input File Format & Grammar Specification](#4-input-file-format--grammar-specification)
   - [Grammar & Rules](#grammar--rules)
   - [Validation & Error Standard](#validation--error-standard)
5. [Codebase Architecture & Component Design](#5-codebase-architecture--component-design)
   - [`pkg/antfarm` — Domain Structs](#pkgantfarm--domain-structs)
   - [`pkg/parser` — Validation & Input Engine](#pkgparser--validation--input-engine)
   - [`pkg/solver` — Graph Pathfinder & Ant Simulator](#pkgsolver--graph-pathfinder--ant-simulator)
   - [`pkg/visualizer` — Terminal Visualizer](#pkgvisualizer--terminal-visualizer)
   - [`web/` & `cmd/web-visualizer` — Go-Backend Powered Web Interface](#web--cmdweb-visualizer--go-backend-powered-web-interface)
6. [Detailed Example Walkthroughs](#6-detailed-example-walkthroughs)
   - [Line-by-Line Breakdown: Example Map](#line-by-line-breakdown-example-map)
   - [Topology Analysis: Examples 04, 06, and 07](#topology-analysis-examples-04-06-and-07)
7. [Edge Cases & Error Handling Matrix](#7-edge-cases--error-handling-matrix)
8. [Build, Test, and Execution Guide](#8-build-test-and-execution-guide)

---

## 1. Executive Overview & Objectives

`lem-in` is a digital ant colony simulation and pathfinding engine written in Go. The goal of the program is to guide $N$ ants from a designated starting room (`##start`) to a destination exit room (`##end`) through a network of connected rooms and tunnels in the **fewest possible turns**.

### Core Problem Statement
Given an ant colony described by:
- Number of ants ($N \ge 1$)
- A set of rooms with $(X, Y)$ integer coordinates
- A set of undirected tunnels connecting pairs of rooms

Find the optimal set of node-disjoint paths and schedule ant movements turn-by-turn such that all $N$ ants reach `##end` in the minimum number of simulation steps.

---

## 2. Domain Rules & Constraints

1. **Ant Population**: At the start of turn 1, all $N$ ants reside in room `##start`.
2. **Goal**: Bring all $N$ ants to room `##end` with the minimum total turns.
3. **Node Capacity (Room Constraint)**:
   - Intermediate rooms can hold **at most 1 ant at a time**.
   - `##start` and `##end` can hold an **unlimited** number of ants simultaneously.
4. **Edge Capacity (Tunnel Constraint)**:
   - Each tunnel connects exactly two rooms.
   - Each tunnel can be traversed by **at most 1 ant per turn**.
5. **Simultaneous Movement**:
   - Multiple ants can move in the same turn, provided they move along valid tunnels and land in empty rooms (or rooms freed during the exact same turn).
   - An ant can move into a room if the ant currently occupying that room leaves it during the same turn (single-file pipelining).
6. **No Room Names with Invalids**:
   - Room names must **not** start with the letter `L` or `#`.
   - Room names must **not** contain spaces.
   - Room coordinates must be valid `int` values.
7. **Error Requirement**:
   - Any invalid input format, missing start/end rooms, self-loops, duplicate rooms, duplicate coordinates, or disconnected graphs must return `ERROR: invalid data format` (with optional descriptive details) and exit cleanly with exit code 1.

---

## 3. Graph Theory & Algorithmic Concepts

### Flow Networks & Node-Disjoint Paths
An ant colony is modeled as an undirected graph $G = (V, E)$. Because intermediate rooms can contain at most 1 ant at a time, we need **node-disjoint paths** (paths that share no intermediate vertices).

Simple edge-disjoint paths are insufficient because two edge-disjoint paths might intersect at an intermediate room, causing a bottleneck collision.

### Node Splitting Technique
To enforce node capacity constraints using standard edge-capacity max-flow algorithms (Edmonds-Karp / Ford-Fulkerson), we transform the graph via **Node Splitting**:

1. For every intermediate room $v \in V \setminus \{\text{start}, \text{end}\}$, split $v$ into two nodes: $v_{\text{in}}$ and $v_{\text{out}}$.
2. Add a directed edge $v_{\text{in}} \to v_{\text{out}}$ with capacity $1$.
3. For every original undirected tunnel $(u, v) \in E$:
   - Add directed edge $u_{\text{out}} \to v_{\text{in}}$ with capacity $1$.
   - Add directed edge $v_{\text{out}} \to u_{\text{in}}$ with capacity $1$.
4. `##start` and `##end` remain unsplit (infinite node capacity).

```text
Original Room V:       ( u ) <=====================> ( v )
                           
Transformed Graph:    (u_out) ------[cap=1]-----> (v_in)
                         |                          |
                       [cap=1]                    [cap=1]
                         v                          v
                      (u_in)  <-----[cap=1]------ (v_out)
```

This transformation guarantees that any flow through room $v$ must traverse $v_{\text{in}} \to v_{\text{out}}$ (capacity 1), enforcing that at most 1 path can use room $v$ simultaneously.

### Edmonds-Karp / Suurballe's Algorithm
The solver iteratively finds augmenting paths in the residual graph:

1. **Augmenting Path Discovery**: Uses Breadth-First Search (BFS) on the residual capacity graph to find shortest augmenting paths in terms of edge count.
2. **Flow Application**: Reverses flow along chosen augmenting paths, allowing back-tracking and path cancellation.
3. **Disjoint Path Extraction**: Traverses positive flow edges deterministically from `##start` to `##end` to extract clean node-disjoint routes.
4. **Turn Evaluation**: After extracting $k$ disjoint paths, the solver evaluates the total turns required for $N$ ants using `CalculateTurns`. It tracks and selects the path combination that minimizes total turns.

### Turn Calculation & Ant Allocation Math
Given a set of $k$ disjoint paths $P_1, P_2, \dots, P_k$ with lengths $L_1 \le L_2 \le \dots \le L_k$ (where $L_i$ is the number of hops/nodes in path $P_i$):

To assign $N$ ants across the $k$ paths to minimize turns, each ant $a \in \{1, \dots, N\}$ is greedily assigned to path $i$ that minimizes the turn arrival cost:
$$\text{Cost}(i) = L_i + \text{ants}_i$$

After allocating all $N$ ants across the paths:
$$\text{Turns}(i) = L_i + \text{ants}_i - 1 \quad (\text{for } \text{ants}_i > 0)$$
$$\text{Total Turns} = \max_{i : \text{ants}_i > 0} \left( L_i + \text{ants}_i - 1 \right)$$

This exact greedy allocation ensures that longer paths are only utilized when doing so reduces the overall turn count compared to crowding shorter paths.

---

## 4. Input File Format & Grammar Specification

### Grammar & Rules
The input file follows a strict line-based format:

```text
<number_of_ants>
<room_definitions>
<tunnel_definitions>
```

1. **Ant Count**: First non-comment, non-empty line must be a positive integer $N > 0$.
2. **Room Definitions**: `name coord_x coord_y`
   - `name`: String containing no spaces, not starting with `L` or `#`.
   - `coord_x`, `coord_y`: Integers (can be negative, e.g., `-5`).
3. **Directives**:
   - `##start`: The next valid room line defines the start room.
   - `##end`: The next valid room line defines the end room.
   - `#comment`: Any line starting with a single `#` (not `##start` or `##end`) is ignored.
4. **Tunnel Definitions**: `room1-room2`
   - Connects `room1` and `room2`. Must reference previously defined rooms. Self-loops (`room1-room1`) or duplicate tunnels are forbidden.

### Validation & Error Standard
If any rule is violated, the program outputs:
```text
ERROR: invalid data format
```
(Optionally appended with descriptive error context) and exits with exit code `1`.

---

## 5. Codebase Architecture & Component Design

```text
lem-in/
├── cmd/
│   ├── lem-in/            # Main CLI entrypoint
│   ├── visualizer/        # Terminal visualizer entrypoint
│   └── web-visualizer/    # Go HTTP web visualizer server
├── pkg/
│   ├── antfarm/           # Core domain models
│   ├── parser/            # Input parsing & validation engine
│   ├── solver/            # Flow network solver & ant simulator
│   └── visualizer/        # Terminal ANSI renderer
├── web/                   # Web assets
│   ├── index.html         # HTML UI template
│   └── static/
│       ├── css/style.css  # Minimalistic, accessible CSS styling
│       └── js/visualizer.js # Canvas renderer fetching Go API
├── testdata/              # Test maps & bad example cases
└── README.md              # Project usage documentation
```

### `pkg/antfarm` — Domain Structs
- `Room`: Name, coordinates $(X, Y)$, `IsStart`, `IsEnd`.
- `Farm`: Number of ants, pointers to `Start` and `End` rooms, `Rooms` map, `Links` map, `RawInput` string.
- `Path`: Slice of room names `[]string`.

### `pkg/parser` — Validation & Input Engine
- **Field-based Detection**: Distinguishes room definitions (`len(fields) == 3`) from tunnel links (`len(fields) == 1 && contains '-'`). This cleanly parses rooms with negative coordinates (e.g., `room1 0 -5`) without misidentifying them as tunnels.
- **Hyphenated Room Tunnel Lookup**: Searches `farm.Rooms` for candidate splits `u` and `v` around hyphens, allowing rooms containing hyphens (e.g., `room-1-end`) to be parsed correctly.
- **Strict State Guard**: Validates that `##start` and `##end` directives are immediately followed by room lines, and checks for duplicate room names, duplicate coordinates, duplicate tunnels, self-loops, and invalid ant counts.

### `pkg/solver` — Graph Pathfinder & Ant Simulator
- `suurballe.go`:
  - `FindOptimalPaths(*Farm)`: Constructs the split-node flow graph, runs BFS augmenting path discovery, applies flow, extracts node-disjoint paths, and picks the path set minimizing `CalculateTurns`.
  - `CalculateTurns()`: Computes exact turns via greedy ant allocation.
  - Deterministic BFS extraction over `fg.adj[curr]` slices (avoids non-deterministic Go map iteration).
- `simulation.go`:
  - `DispatchAndSimulate(paths, totalAnts)`: Runs the turn-by-turn ant movement pipeline. Advances active ants single-file and dispatches new ants onto available paths each turn. Returns output formatted as `Lx-y Lz-w`.

### `pkg/visualizer` — Terminal Visualizer
- Terminal ANSI visualizer rendering auto-scaled grid frames with step timing.

### `web/` & `cmd/web-visualizer` — Go-Backend Powered Web Interface
- Go server (`cmd/web-visualizer/main.go`) exposes `/api/maps` and `/api/simulate`.
- It executes Go `parser` and `solver` on the backend, returning JSON containing rooms, links, and calculated turn movements to `web/static/js/visualizer.js` for canvas rendering.
- UI styling is minimal, muted, and colorblind-friendly.

---

## 6. Detailed Example Walkthroughs

### Line-by-Line Breakdown: Example Map

Consider the following execution:

```text
3
2 5 0
##start
0 1 2
##end
1 9 2
3 5 4
0-2
0-3
2-1
3-1
2-3

L1-2 L2-3
L1-1 L2-1 L3-2
L3-1
```

#### Input Map Breakdown

| Input Line | Meaning |
| :--- | :--- |
| `3` | Total number of ants = 3. |
| `2 5 0` | Room `"2"` at coordinates $(5, 0)$. |
| `##start` | Directive: Next room is the starting room. |
| `0 1 2` | Room `"0"` at coordinates $(1, 2)$ — **`##start`**. |
| `##end` | Directive: Next room is the exit room. |
| `1 9 2` | Room `"1"` at coordinates $(9, 2)$ — **`##end`**. |
| `3 5 4` | Room `"3"` at coordinates $(5, 4)$. |
| `0-2` | Tunnel between room `0` and room `2`. |
| `0-3` | Tunnel between room `0` and room `3`. |
| `2-1` | Tunnel between room `2` and room `1`. |
| `3-1` | Tunnel between room `3` and room `1`. |
| `2-3` | Tunnel between room `2` and room `3`. |

#### Simulation Output Breakdown

The solver identifies 2 node-disjoint paths of length 2:
- **Path A**: `0 -> 2 -> 1`
- **Path B**: `0 -> 3 -> 1`

Movement steps (`L[AntID]-[Room]`):

1. **Turn 1: `L1-2 L2-3`**
   - Ant 1 (`L1`) moves from `##start` (`0`) to room `2`.
   - Ant 2 (`L2`) moves from `##start` (`0`) to room `3`.
   - *(Ant 3 waits at `0` as both paths are occupied).*

2. **Turn 2: `L1-1 L2-1 L3-2`**
   - Ant 1 (`L1`) moves from room `2` into `##end` (`1`). *(Arrived!)*
   - Ant 2 (`L2`) moves from room `3` into `##end` (`1`). *(Arrived!)*
   - Ant 3 (`L3`) dispatches from `##start` (`0`) into room `2` (now empty).

3. **Turn 3: `L3-1`**
   - Ant 3 (`L3`) moves from room `2` into `##end` (`1`). *(Arrived!)*

---

### Topology Analysis: Examples 04, 06, and 07

In maps `example04.txt` (9 ants), `example06.txt` (100 ants), and `example07.txt` (1,000 ants), the graph topology is:

```text
          [dinish] -------> [jimYoung]
         /                             \
[richard] -------> [erlich] -----------> [peter]
         \                             /
          ---------> [gilfoyle] -------
```

#### Why ants DO NOT use the middle node (`erlich`):

1. **Ingress Bottleneck at `##end` (`peter`)**:
   `peter` has only **2 incoming tunnels** (`gilfoyle-peter` and `jimYoung-peter`). Thus, at most **2 ants can enter `peter` per turn**, limiting the maximum node-disjoint paths to **2**.

2. **Node Contention**:
   Any path through `erlich` must exit to either `gilfoyle` or `jimYoung`.
   - Routing `richard -> erlich -> gilfoyle -> peter` (3 hops) consumes `gilfoyle`, blocking the shorter 2-hop path (`richard -> gilfoyle -> peter`).
   - Routing `richard -> erlich -> jimYoung -> peter` (3 hops) consumes `jimYoung`, blocking the 3-hop path (`richard -> dinish -> jimYoung -> peter`).

3. **Path Selection**:
   The solver selects the optimal 2-path set:
   - **Path 1**: `richard -> gilfoyle -> peter` (Length 2)
   - **Path 2**: `richard -> dinish -> jimYoung -> peter` (Length 3)

Using `erlich` would either increase overall path length or conflict with shorter paths, so it is omitted by Suurballe's algorithm.

---

## 7. Edge Cases & Error Handling Matrix

| Edge Case Scenario | Expected Behavior / Output |
| :--- | :--- |
| Negative Room Coordinates (`roomA 0 -5`) | Parsed correctly as valid integers. |
| Hyphenated Room Names (`room-1 0 0`) | Parsed correctly using room table matching. |
| Duplicate Room Name (`A 0 0`, `A 1 1`) | `ERROR: invalid data format, duplicate room name` |
| Duplicate Room Coordinates (`A 0 0`, `B 0 0`) | `ERROR: invalid data format, duplicate room coordinates` |
| Room Name Starting with `L` or `#` | `ERROR: invalid data format, room name cannot start with a # or L` |
| Room Name with Spaces (`room A 0 0`) | `ERROR: invalid data format, room definition format` |
| Ant Count $\le 0$ or Non-Integer | `ERROR: invalid data format, invalid number of ants` |
| Self-Referencing Tunnel (`A-A`) | `ERROR: invalid data format, self referencing tunnel` |
| Duplicate Tunnel (`A-B`, `A-B`) | `ERROR: invalid data format, duplicate tunnel` |
| Tunnel to Unknown Room (`A-Z`) | `ERROR: invalid data format, tunnel links unknown room` |
| Missing `##start` or `##end` | `ERROR: invalid data format, missing ##start and ##end room` |
| No Path Between `##start` and `##end` | `ERROR: invalid data format, no path found between ##start and ##end` |
| `##start` Not Followed by Room Line | `ERROR: invalid data format, ##start or ##end not followed by a room` |

---

## 8. Build, Test, and Execution Guide

### Prerequisites
- Go version 1.18+

### 1. Build Binaries

**Linux / macOS:**
```bash
go build -o lem-in ./cmd/lem-in
go build -o visualizer ./cmd/visualizer
```

**Windows (PowerShell):**
```powershell
go build -o lem-in.exe ./cmd/lem-in
go build -o visualizer.exe ./cmd/visualizer
```

---

### 2. Run CLI Engine

```bash
# Using compiled binary:
./lem-in testdata/example00.txt

# Or direct via go run:
go run ./cmd/lem-in testdata/example00.txt
```

---

### 3. Run Terminal Visualizer

```bash
# Piping lem-in output into terminal visualizer:
./lem-in testdata/example00.txt | ./visualizer
```

---

### 4. Run Interactive Web Visualizer

Start the Go web visualizer server:
```bash
go run ./cmd/web-visualizer
```
Open **`http://localhost:8080`** in your browser.

---

### 5. Run Unit Test Suite

Execute all package tests with verbose logs:
```bash
go test -v ./pkg/...
```
