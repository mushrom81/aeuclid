# aeuclid
Basic framework for creating a voxel-based non-euclidian world

This module implements two main objects: *Room*s and *Orientation*s

### Room
This is the more basic of the type of the two, basically acting as a 3d array.
It also contains a list of connection *Orientation*s, which dictate how the each *Room* connects to its neighbors.
This will be explained further later.

### Orientation
When designing any sort of framework for some kind of non-euclidian world, we need to think about how position will be communicated.
The first thing that came to my mind was to just include a pointer to the current room alongside the standard xyz coordinates.
Later, I realized that it also made sense to include some sense of direction, since the concept of absolute north is not preserved in non-euclidian spaces.
On top of that, these souped-up location structs also make the transformation between different rooms a lot simpler.

## How it works
### Using coordinate
First off, we have the rooms, which are just 3d arrays. The coordinates relative to the room just directly index into the array.
The direction doesn't affect how the coordinates are relative to the array, but it affects what coordinates relative to *it* represent, similarly to how the direction you are facing changes what is in front of you.

We can compound two of these coordinate-rotations together by taking one as relative to the other. This means taking the position and rotation of the first and treating it as the 0,0 for the second one.
This compound will be relative to the room the first coordinate is relative to.
For example, when the rotation of the first pair is zero, the compound is simply the addition of the xyz components.

When we want to describe the way two rooms are joined, we represent the one-directional connection with a single relative coordinate-rotation (Orientation).
Conceptually, this Orientation can be thought of as the origin of the initial room in terms of the coordinates of the connected room.
When it comes to "moving" an Orientation from one room to the other, one simply has to compound this connection rotation with the Orientation to be moved to get the correct coordinate in terms of the new room.

### Getting a relative tile
When it comes to getting the value of a tile relative to some arbitrary position (as in seeing), we have to be careful. Unless the position is entirely along a single axis (forward, forward, forward) there will be several permutations of moves that end up at the same position (up, left vs. left, up).
Since these are not guaranteed to be equivalent, we have to check through all of them and ensure that they all lead to the same place.
If they are not all the same, then we can't give a single definitive position, and so we say it is ambiguous.
The one exception to this rule is when one of the resulting positions is out-of-bounds.
In that case, we can ignore the out-of-bounds position since it never makes sense to access such a position intentionally.
It is also worth noting that once a position is either out-of-bounds or ambiguous, it cannot ever be translated back into a valid one, since the question of which translations would do such a thing doesn't really make sense in a non-euclidian landscape.
