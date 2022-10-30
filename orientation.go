package aeuclid

import (
	"errors"
)

// Cardinal directions away from the origin
// 0-3 are clockwise rotations from +x about +z
type Dir int

const (
	X_PLUS  Dir = 0
	Y_PLUS  Dir = 1
	X_MINUS Dir = 2
	Y_MINUS Dir = 3
	Z_PLUS  Dir = 4
	Z_MINUS Dir = 5
)

// Orientations represent a position and rotation relative to some room
// The rotation is relative to the position, not the origin of the room
type Orientation struct {
	room    *Room
	x, y, z int
	r       Dir
}

// Constants for orientations that resolve to T_AMBG and T_OOB respectively
// These are technically mutable, but they are not exported and none of the above code alters them
// Go does not support constant structs, and I don't like the idea of using a function
var r_AMBG = *NewRoom(0, 0, 0)
var o_AMBG Orientation = Orientation{&r_AMBG, 0, 0, 0, X_PLUS} // HAAAATE this
// See, what I'm doing here is I'm calling NewRoom() within the struct declaration
// so that the pointer is not nil, but I can guarentee that it will be unique to this use case
// So, its basically like a second nil
// but the way I've done it is utterly unreadable
// Thank goodness it doesn't have to be exported
var o_OOB Orientation = Orientation{nil, 0, 0, 0, X_PLUS}

func NewOrientation(room *Room, x, y, z int, r Dir) Orientation {
	return NewOrientation(room, x, y, z, r)
}

func (initialOri Orientation) Step(steps ...Dir) Orientation {
	if initialOri.IsAMBG() {
		return o_AMBG
	}
	if initialOri.IsOOB() {
		return o_OOB
	}
	if len(steps) == 0 {
		return initialOri
	}
	if len(steps) == 1 {
		return initialOri.unitStep(steps[0])
	}

	finalOri := o_OOB
	// Default value for bool is false
	// Basically, stepping is not a commutative operation so we have to try all possible combinations of moves
	// What we do here is iterate throught every possible step that we could take first, and then recurse so the problem solves itself
	dirsSteppedFirst := make([]bool, 6)

	for i := range steps {
		if !dirsSteppedFirst[steps[i]] {
			steps[0], steps[i] = steps[i], steps[0]
			// Recursion is really epic and cool and it is one of the best things that you can possibly do
			possibleFinalDir := initialOri.unitStep(steps[0]).Step(steps[1:]...)
			steps[0], steps[i] = steps[i], steps[0]

			if possibleFinalDir.IsAMBG() {
				return o_AMBG
			}
			if !possibleFinalDir.IsOOB() {
				if !finalOri.IsOOB() && finalOri != possibleFinalDir {
					return o_AMBG
				}
				finalOri = possibleFinalDir
			}

			dirsSteppedFirst[steps[i]] = true
		}
	}
	return finalOri
}

// One step along an axis direction
func (ori Orientation) unitStep(dir Dir) Orientation {
	if ori.IsAMBG() {
		return o_AMBG
	}
	if ori.IsOOB() {
		return o_OOB
	}

	switch rotate(ori.r, dir) {
	case X_PLUS:
		ori.x++
	case Y_PLUS:
		ori.y++
	case X_MINUS:
		ori.x--
	case Y_MINUS:
		ori.y--
	case Z_PLUS:
		ori.z++
	case Z_MINUS:
		ori.z--
	}

	if !ori.IsOOB() {
		return ori
	}

	finalOri := o_OOB
	for _, door := range ori.room.connections {
		possibleFinalOri := door.Plus(ori)
		if !possibleFinalOri.IsOOB() {
			if !finalOri.IsOOB() && finalOri != possibleFinalOri {
				return o_AMBG
			}
			finalOri = possibleFinalOri
		}
	}
	return finalOri
}

// Assume b is relative to a and compute their compounded orientation
func (a Orientation) Plus(b Orientation) Orientation {
	switch a.r {
	case 1:
		b.x, b.y = b.y, -b.x
	case 2:
		b.x, b.y = -b.x, -b.y
	case 3:
		b.x, b.y = -a.y, b.x
	}
	a.x += b.x
	a.y += b.y
	a.z += b.z
	a.r = rotate(a.r, b.r)
	return a
}

// ori.Plus(InverseOf(ori)) == Orientation{ori.room, 0, 0, 0, X_PLUS}
// or
// ori.Plus(ori.Inverse()) == Orientation{ori.room, 0, 0, 0, X_PLUS}
// ?
func InverseOf(ori Orientation) Orientation {
	return Orientation{ori.room, 0, 0, 0, X_MINUS}.Plus(ori)
}

// Rotates a by b, assuming b is a rotation awey from X_PLUS
// This is kinda goofy, but we don't end up using vertical rotation so its mostly there for consistancy
func rotate(a Dir, b Dir) Dir {
	aIsVertical := a >= 4
	bIsVertical := b >= 4
	if !aIsVertical && !bIsVertical {
		a += b // mod 4
		a %= 4
		if a < 0 {
			a += 4
		}
		return a
	} else if aIsVertical {
		return a
	} else if bIsVertical {
		return b
	}
	// a and b are both either Z_PLUS or Z_MINUS
	if a == b {
		return X_MINUS
	} else {
		return X_PLUS
	}
}

func (ori Orientation) Spin(rot Dir) Orientation {
	ori.r = rotate(ori.r, rot)
	return ori
}

func (loc Orientation) Get() (Tile, error) {
	if loc.IsAMBG() {
		return 0, errors.New(".Get(): accessing ambiguous tile")
	}
	if loc.IsOOB() {
		return 0, errors.New(".Get(): accessing out-of-bounds tile")
	}
	index := loc.x
	index += loc.y * (loc.room.dimX)
	index += loc.z * (loc.room.dimX * loc.room.dimY)
	return loc.room.terrain[index], nil
}

func (loc Orientation) Set(value Tile) error {
	if loc.IsAMBG() {
		return errors.New(".Set(): accessing ambiguous tile")
	}
	if loc.IsOOB() {
		return errors.New(".Set(): accessing out-of-bounds tile")
	}
	index := loc.x
	index += loc.y * (loc.room.dimX)
	index += loc.z * (loc.room.dimX * loc.room.dimY)
	loc.room.terrain[index] = value
	return nil
}

// This is a little confusing, since IsOOB() says something much more general about an Orientation
// while IsAMBG() is merely a value comparison

func (loc Orientation) IsAMBG() bool {
	return loc.room == &r_AMBG
}

func (loc Orientation) IsOOB() bool {
	if loc.IsAMBG() {
		return false
	}
	if loc.room == nil {
		return true
	}
	if loc.x < 0 || loc.room.dimX <= loc.x ||
		loc.y < 0 || loc.room.dimY <= loc.y ||
		loc.z < 0 || loc.room.dimZ <= loc.z {
		return true
	}
	return false
}
