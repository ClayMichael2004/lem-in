//validates all constraints
//positive ant count, unique room names and coordinates
//room name must not start with L or # and must contain no spaces.
//strictly validates ##start and ##end directives
//ignore #... while validating standard tunnels.

package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"lem-in/pkg/antfarm"
)

// ParseFile parses a filepath and returns a Farm or a descriptive error
func ParseFile(filePath string) (*antfarm.Farm, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ERROR: invalid data format, cannot open file: %w", err)
	}
	defer file.Close()
	return ParseReader(file)
}

func ParseReader(r io.Reader) (*antfarm.Farm, error) {
	scanner := bufio.NewScanner(r)
	farm := antfarm.NewFarm()

	var (
		rawLines     []string
		antsParsed   bool
		nextIsStart  bool
		nextIsEnd    bool
		coordsSeen   = make(map[string]bool)
		readingLinks bool
		linkSeen     = make(map[string]bool)
	)

	for scanner.Scan() {
		line := scanner.Text()
		rawLines = append(rawLines, line)
		trimmed := strings.TrimSpace(line)

		// ignore empty lines
		if trimmed == "" {
			continue
		}

		// handle comments and commands
		if strings.HasPrefix(trimmed, "#") {
			if trimmed == "##start" {
				if farm.Start != nil || nextIsStart {
					return nil, errors.New("ERROR: invalid data format, multiple ##start commands")
				}
				if nextIsEnd {
					return nil, errors.New("ERROR: invalid data format, ##start immediately after ##end")
				}
				nextIsStart = true
			} else if trimmed == "##end" {
				if farm.End != nil || nextIsEnd {
					return nil, errors.New("ERROR: invalid data format, multiple ##end commands")
				}
				if nextIsStart {
					return nil, errors.New("ERROR: invalid data format, ##end immediately after ##start")
				}
				nextIsEnd = true
			}
			// other comments beginning with # are ignored
			continue
		}

		// parsing the number of ants
		if !antsParsed {
			ants, err := strconv.Atoi(trimmed)
			if err != nil || ants <= 0 {
				return nil, errors.New("ERROR: invalid data format, invalid number of ants")
			}
			farm.Ants = ants
			antsParsed = true
			continue
		}

		fields := strings.Fields(trimmed)
		isRoomLine := len(fields) == 3
		isLinkLine := strings.Contains(trimmed, "-") && len(fields) == 1

		if isRoomLine {
			if readingLinks {
				return nil, errors.New("ERROR: invalid data format, room definition after links")
			}

			roomName := fields[0]
			if strings.HasPrefix(roomName, "L") || strings.HasPrefix(roomName, "#") {
				return nil, errors.New("ERROR: invalid data format, room name cannot start with a # or L")
			}
			if _, exists := farm.Rooms[roomName]; exists {
				return nil, errors.New("ERROR: invalid data format, duplicate room name")
			}

			x, errX := strconv.Atoi(fields[1])
			y, errY := strconv.Atoi(fields[2])
			if errX != nil || errY != nil {
				return nil, errors.New("ERROR: invalid data format, room coordinates must be integers")
			}

			coordKey := fmt.Sprintf("%d,%d", x, y)
			if coordsSeen[coordKey] {
				return nil, errors.New("ERROR: invalid data format, duplicate room coordinates")
			}
			coordsSeen[coordKey] = true

			room := &antfarm.Room{
				Name:    roomName,
				X:       x,
				Y:       y,
				IsStart: nextIsStart,
				IsEnd:   nextIsEnd,
			}

			if nextIsStart {
				farm.Start = room
				nextIsStart = false
			}
			if nextIsEnd {
				farm.End = room
				nextIsEnd = false
			}

			farm.Rooms[roomName] = room
		} else if isLinkLine {
			readingLinks = true

			if nextIsStart || nextIsEnd {
				return nil, errors.New("ERROR: invalid data format, ##start or ##end not followed by a room")
			}

			var u, v string
			found := false

			for i := 1; i < len(trimmed)-1; i++ {
				if trimmed[i] == '-' {
					candU := trimmed[:i]
					candV := trimmed[i+1:]
					if _, okU := farm.Rooms[candU]; okU {
						if _, okV := farm.Rooms[candV]; okV {
							u, v = candU, candV
							found = true
							break
						}
					}
				}
			}

			if !found {
				parts := strings.Split(trimmed, "-")
				if len(parts) != 2 {
					return nil, errors.New("ERROR: invalid data format, invalid link/tunnel format")
				}
				u, v = parts[0], parts[1]
				if _, existsU := farm.Rooms[u]; !existsU {
					return nil, errors.New("ERROR: invalid data format, tunnel links unknown room")
				}
				if _, existsV := farm.Rooms[v]; !existsV {
					return nil, errors.New("ERROR: invalid data format, tunnel links unknown room")
				}
			}

			if u == v {
				return nil, errors.New("ERROR: invalid data format, self referencing tunnel")
			}

			linkKey1 := u + "-" + v
			linkKey2 := v + "-" + u
			if linkSeen[linkKey1] || linkSeen[linkKey2] {
				return nil, errors.New("ERROR: invalid data format, duplicate tunnel")
			}
			linkSeen[linkKey1] = true
			linkSeen[linkKey2] = true

			farm.AddLink(u, v)
		} else {
			return nil, errors.New("ERROR: invalid data format, unrecognized line format")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ERROR: invalid data format, scanner error: %w", err)
	}

	if !antsParsed {
		return nil, errors.New("ERROR: invalid data format, missing ant count")
	}
	if farm.Start == nil || farm.End == nil {
		return nil, errors.New("ERROR: invalid data format, missing ##start and ##end room")
	}
	if nextIsStart || nextIsEnd {
		return nil, errors.New("ERROR: invalid data format, ##start or ##end not followed by a room")
	}

	farm.RawInput = strings.Join(rawLines, "\n")
	return farm, nil
}