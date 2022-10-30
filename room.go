package aeuclid

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Tile int

// A room is a rectangular prism, represented by a 3d array and using local coordinates
// Ex: a rubik's cube is a 3x3x3 room
// Rooms can be connected together through the use of doors
type Room struct {
	terrain          []Tile        // The array is 1d, but we use modulos to treat it as 3d
	dimX, dimY, dimZ int           // Dimensions of the 3d array
	connections      []Orientation // Orientation of the new room's origin relative to the old room
}

// Create a blank room of specified dimensions
func NewRoom(dimX, dimY, dimZ int) *Room {
	var room Room
	room.dimX, room.dimY, room.dimZ = dimX, dimY, dimZ
	room.terrain = make([]Tile, dimX*dimY*dimZ)
	room.connections = make([]Orientation, 0)
	return &room
}

// Calculate and add door and its inverse to pair of rooms
func (room1 *Room) AddDoor(origin2 Orientation) {
	room2 := origin2.room
	origin1 := InverseOf(origin2)
	room1.connections = append(room1.connections, origin2)
	room2.connections = append(room2.connections, origin1)
}

// These functions aren't really important
// They just implement serialization and deserialization
// into and out of a format I made up with in about five minutes
// Its probably better to just think of them as black box functions
// and take them at face value

func SerializeWorld(world []Room) []byte {
	str := fmt.Sprintf("%d\n\n", len(world))
	for i := range world {
		for _, door := range world[i].connections {
			str += fmt.Sprintf("%d, %d, %d, %d, %d\n",
				door.x, door.y, door.z, int(door.r),
				reverseIndex(world, door.room))
		}
		str += fmt.Sprintf("%d, %d, %d\n",
			world[i].dimX, world[i].dimY, world[i].dimZ)
		for index, tile := range world[i].terrain {
			if index%world[i].dimX != 0 {
				str += ", "
			}
			str += fmt.Sprintf("%d", tile)
			if (index+1)%world[i].dimX == 0 {
				str += "\n"
				if (index+1)%(world[i].dimY) == 0 {
					str += "\n"
				}
			}
		}
	}
	buf := make([]byte, 0)
	buf = append(buf, []byte(str)...)
	return buf
}

func DeserializeWorld(serial io.Reader) ([]Room, error) {
	data := bufio.NewScanner(serial)

	line, err := getNextField(data)
	if err != nil {
		return nil, err
	}
	roomCount, err := strconv.Atoi(strings.TrimSpace(line[0]))
	if err != nil {
		return nil, err
	}
	world := make([]Room, roomCount)

	for i := range world {
		world[i].connections = make([]Orientation, 0)
		for {
			line, err := getNextField(data)
			if err != nil {
				return nil, err
			}
			x, err := strconv.Atoi(strings.TrimSpace(line[0]))
			if err != nil {
				return nil, err
			}
			y, err := strconv.Atoi(strings.TrimSpace(line[1]))
			if err != nil {
				return nil, err
			}
			z, err := strconv.Atoi(strings.TrimSpace(line[2]))
			if err != nil {
				return nil, err
			}
			if len(line) == 3 {
				world[i].dimX, world[i].dimY, world[i].dimZ = x, y, z
				break
			}
			r, err := strconv.Atoi(strings.TrimSpace(line[3]))
			if err != nil {
				return nil, err
			}
			room, err := strconv.Atoi(strings.TrimSpace(line[4]))
			if err != nil {
				return nil, err
			}
			world[i].connections = append(world[i].connections, Orientation{&world[room], x, y, z, Dir(r)})
		}
		world[i].terrain = make([]Tile, world[i].dimX*world[i].dimY*world[i].dimZ)
		for z := 0; z < world[i].dimZ; z++ {
			for y := 0; y < world[i].dimY; y++ {
				line, err := getNextField(data)
				if err != nil {
					return nil, err
				}
				for x := 0; x < world[i].dimX; x++ {
					value, err := strconv.Atoi(strings.TrimSpace(line[x]))
					world[i].terrain[x+y*(world[i].dimX)+z*(world[i].dimX*world[i].dimY)] = Tile(value)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return world, nil
}

// This is basically a macro to get the next csv value out of the file
func getNextField(scanner *bufio.Scanner) ([]string, error) {
	scanner.Scan()
	if scanner.Text() == "" {
		scanner.Scan()
	}
	return strings.Split(scanner.Text(), ","), scanner.Err()
}

// Return the first index of the occurence of a value in an array
// I'm using any's here because I have the delusion that this function
// could possibly be used in a larger scope than this single file
// -------------- []array -- *value
func reverseIndex(array any, value any) int {
	for key := range array.([]any) {
		if array.([]any)[key] == value {
			return key
		}
	}
	return -1
}
