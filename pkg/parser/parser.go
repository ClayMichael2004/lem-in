//validates all constraints
//positive and count, unique room names and coordinates
//room number must not start with L or # and must contain  no space.
//stricly validates ##start and ##end directives
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

//parses a filepath and returns a Farm or a descriptive error
func ParseFile(filePath string) (*antfarm.Farm, error){
	file, error:= os.Open(filePath)
	if error!=nil{
		return nil, fmt.Errorf("ERROR: invalid data format, cannot open file: %w", error)
	}
	defer file.Close()
	return ParseReader(file)
}

func ParseReader(r io.Reader) (*antfarm.Farm, error){
	scanner:= bufio.NewScanner(r)
	farm:= antfarm.NewFarm()

	var(
		rawLines  []string
		antsParsed bool
		nextIsStart bool
		nextIsEnd   bool
		coordsSeen  = make(map[string]bool)
		readingLinks  bool
	)

	for scanner.Scan(){
		line:= scanner.Text()
		rawLines = append(rawLines, line)
		trimmed:= strings.TrimSpace(line)

		//ignore empty lines
		if trimmed==""{
			continue
		}

		//handle comments and commands
		if strings.HasPrefix(trimmed, "#"){
			if trimmed == "##start"{
				if farm.Start !=nil{
					return nil, errors.New("ERROR: invalid data format, multiple ##start commands")
				}
				nextIsStart=true
			}else if trimmed=="##end"{
				if farm.End !=nil{
					return nil, errors.New("ERROR: invalid data format, multiple ##end commands")
				}
				nextIsEnd=true
			}
			//other commenst beginning with # are ignored
			continue
		}
		//parsing the number of ants
		if !antsParsed{
			ants, err:= strconv.Atoi(trimmed)
			if err != nil || ants <=0{
				return nil, errors.New("ERROR: invalid data format, invalid number of ants")
			}
			farm.Ants = ants
			antsParsed = true
			continue
		}

		//detect if we moved from rooms to links or viceversa
		if strings.Contains(trimmed, "-"){
			readingLinks=true
		}
		if !readingLinks{
			//parse room : name x y
			parts:= strings.Fields(trimmed)
			if len(parts)!=3{
				return nil, errors.New("ERROR: invalid data format, room definition format")
			}

			roomName:= parts[0]
			if strings.HasPrefix(roomName, "L") || strings.HasPrefix(roomName, "#"){
				return nil, errors.New("ERROR: invalid data format, room name cannot start with a # or L")
			}
			if _, exists:= farm.Rooms[roomName]; exists{
				return nil, errors.New("ERROR: invalid data format, duplicate room name")
			}

			x, errX:= strconv.Atoi(parts[1])
			y, errY:= strconv.Atoi(parts[2])
			if errX!= nil || errY!=nil{
				return nil, errors.New("ERROR: invalid data format, room coordinates must be integers")
			}

			coordKey:= fmt.Sprintf("%d, %d", x ,y)
			if coordsSeen[coordKey]{
				return nil, errors.New("ERROR: invalid data format, duplicate room cooordinates")
			}

			coordsSeen[coordKey]=true

			room:= &antfarm.Room{
				Name:   roomName,
				X:     x,
				Y:     y,
				IsStart: nextIsStart,
				IsEnd: nextIsEnd,
			}

			if nextIsStart{
				farm.Start=room
				nextIsStart=false
			}
			if nextIsEnd{
				farm.End=room
				nextIsEnd=false
			}

			farm.Rooms[roomName] = room
		}else{
			//parse tunnel/link room1-room2
			parts:= strings.Split(trimmed, "-")
			if len(parts)!=2{
				return nil, errors.New("ERROR: invalid data format, invalid link/tunnel format")
			}

			u, v:= parts[0], parts[1]
			if u==v{
				return nil, errors.New("ERROR: invalid data format, self referencing tunnel")
			}
			if _, existsU:= farm.Rooms[u]; !existsU{
				return nil, errors.New("ERROR: invalid data format, tunnel links unknown room")
			}
			if _, existsV:= farm.Rooms[v]; !existsV{
				return nil, errors.New("ERROR: invalid data format, tunnel links unknown room")
			}
			farm.AddLink(u, v)
		}
	}

	if err:= scanner.Err(); err!=nil{
		return nil, fmt.Errorf("ERROR: invalid data format, scanner error: %w", err)
	}

	if farm.Start==nil || farm.End==nil{
		return nil, errors.New("ERROR: invalid data format, missing ##start and ##end room")
	}

	farm.RawInput=strings.Join(rawLines, "\n")
	return farm, nil
}