package main


import (
	"math/rand"
)

/*
*	exp
*		Standard exponent function
*	parameters:
*		base float64
*		power float 64
*	returns:
*		float64
*/
func exp(base float64, power float64) float64 {
	if power == 0 {
		return 1
	}
	for power > 1 {
		base *= base
		power -= 1
	}
	return base
}

/*
*	expandDirections
*		expands encoded directions object into a slice of directions
*	parameters:
*		encoded int
*	returns:
*		[]Direction
*/
func expandDirections(encoded int) []Direction {
	var decoded []Direction
	if encoded % int(Up) == 0 {
		decoded = []Direction{Up}
	}
	if encoded % int(Left) == 0 {
		decoded = append(decoded, Left)
	}
	if encoded % int(Right) == 0 {
		decoded = append(decoded, Right)
	}
	if encoded % int(Down) == 0 {
		decoded = append(decoded, Down)
	}
	return decoded
}
/*
*	flip
*		Returns the flipped 
*	parameters:
*		start Position
*		dir Direction
*	returns:
*		Position
*/
func flip(dir Direction) Direction {
	switch dir {
	case Up:
		return Down
	case Left:
		return Right
	case Right:
		return Left
	case Down:
		return Up
	default:
		return Down
	}
}

/*
*	move
*		Gets the postition from start after moving in dir direction
*	parameters:
*		start Position
*		dir Direction
*	returns:
*		Position
*/
func move(start Position, dir Direction) Position {
	switch dir {
	case Up:
		return Position{start.Y - 1, start.X}
	case Left:
		return Position{start.Y, start.X - 1}
	case Right:
		return Position{start.Y, start.X + 1}
	case Down:
		return Position{start.Y + 1, start.X}
	default:
		return start
	}
}

/*
*	getNeighbours
*		Gets the postitions from start after moving in dir direction
*	parameters:
*		home Position
*		directions []Direction
*	returns:
*		[]Position
*/
func getNeighbours(home Position, directions []Direction) []Position {
	neighbours := []Position{}
	for _, direction := range directions {
		neighbours = append(neighbours, move(home, direction))
	}
	return neighbours
}

/*
*	rateSquare
*		Recursively rates a square by its child nodes and context in the game 
*	paramaters:
*		pos Position
*		origin Direction
*		distance int
*		depth int
*		length int
*		grownby int
*		health int
*		history []Position{int, int}
*	returns:
*		Rating{float64, int}
*/
func (matrix *Matrix) rateSquare(pos Position, origin Direction, distance int, depth int, length int, grownby int, health int, history []Position) Rating {
	var y, x int = pos.Y, pos.X

	// out of bounds
	if x == -1 || x == matrix.Width || y == -1 || y == matrix.Height {
		return Rating{0, distance}
	}

	// forbidden move
	if matrix.Matrix[y][x].Base == 0 {
		return Rating{0, distance}
	}

	// currently occupied
	eatenOffset := 0
	if matrix.Matrix[y][x].Self {
		eatenOffset = grownby
	}
	if matrix.Matrix[y][x].Tenure + eatenOffset >= distance {
		return Rating{0, distance}
	}

	// occupied by current path
	for h := range history {
		past := &history[h]
		if x == past.X {
			if y == past.Y {
				return Rating{0, distance}
			}
		}
	}

	// set base value
	health -= 1
	base := matrix.Matrix[y][x].Base
	if matrix.Matrix[y][x].Food {
		grownby += 1
		// to promote moderation, 25 <-> 20, 4 <-> 2
		var hungerModifier float64 = 4 / (exp(2, float64(health) / 25))
		base += float64(100 / (distance * distance)) * 4 * hungerModifier
		health = 100
	}

	// return base value (base case)
	if depth == 0 {
		return Rating{base, distance}
	}

	// add current position to history
	history = append(history, Position{y, x})

	// remove last position in history if tenure is up
	if length < depth && len(history) >= length + grownby {
		history = history[1:]
	}

	// continue search (recursive case)
	var directions = expandDirections(210 / int(origin))
 
	var rating Rating

	// modify rating by rating of potential future moves
	for _, direction := range directions {
		node := matrix.rateSquare(
			move(pos, direction),
			flip(direction),
			distance + 1,
			depth - 1,
			length,
			grownby,
			health,
			history,
		)
		rating.Value += base * node.Value / 3
		if node.Distance > rating.Distance {
			rating.Distance = node.Distance
		}
	}

	return rating
}

/*
*	step
*		main logic function that returns calculated approximate best next move
*	paramaters:
*		data Req
*	returns:
*		string
*/
func step(data Req) string {
	bWidth := data.Board.Width
	bHeight := data.Board.Height
	mHead := Position{data.You.Body[0].Y, data.You.Body[0].X}
	mX, mY := mHead.X, mHead.Y
	mLength := len(data.You.Body)

	var directions []Direction

	var x1, y1 int = data.You.Body[1].X, data.You.Body[1].Y

	if mX < x1 {
		directions = expandDirections(210 / int(Right))
	} else if mX > x1 {
		directions = expandDirections(210 / int(Left))
	} else if mY < y1 {
		directions = expandDirections(210 / int(Down))
	} else if mY > y1 {
		directions = expandDirections(210 / int(Up))
	}

	var matrix = Matrix{
		make([][]Square, bHeight),
		bWidth,
		bHeight,
		[]Head{},
		[]Position{},
	}
	var allocation = make([]Square, bHeight * bWidth)
	for i := range matrix.Matrix {
    	matrix.Matrix[i] = allocation[i*bWidth: (i+1)*bWidth]
	}

	// createMatrix
	for y := range matrix.Matrix {
		for x := range matrix.Matrix[y] {
			var v float64 = 1
			var heatmap bool = true

			// Give edge & corner squares a lower base value (and )
			if x == 0 || x == bWidth - 1 {
				v -= 0.25
			} else if heatmap && (y == 2 || y == bHeight - 3) {
				v += 0.25
			}
			if y == 0 || y == bHeight - 1 {
				v -= 0.25
			} else if heatmap && (x == 2 || x == bWidth - 3) {
				v += 0.25
			}

			// Initialize randomModifier
			var randomModifier float64 = 0.1

			// Increase square value by random value if randomModifier > 0
			if randomModifier != 0 {
				v += rand.Float64() * randomModifier
			}

			matrix.Matrix[y][x] = Square{
				Tenure: 0,
				Danger: 0,
				Food: false,
				Self: false,
				Base: v,
			}
		}
	}

	// populateMatrix
	for i := range data.Board.Food {
		food := &data.Board.Food[i]
		matrix.Matrix[food.Y][food.X].Food = true
	}

	// set tenure / matrix's heads
	for i := range data.Board.Snakes {
		snake := &data.Board.Snakes[i]
		id := snake.ID
		head := snake.Body[0]
		oLength := len(snake.Body)

		if id != data.You.ID {
			matrix.Heads = append(matrix.Heads, Head{Position{head.Y, head.X}, oLength})

			// generate squares next to head
			var neighbours []Position
			if (head.X > 0) {
				neighbours = append(neighbours, Position{head.Y, head.X - 1})
			}
			if (head.X < bWidth - 1) {
				neighbours = append(neighbours, Position{head.Y, head.X + 1})
			}
			if (head.Y > 0) {
				neighbours = append(neighbours, Position{head.Y - 1, head.X})
			}
			if (head.Y < bHeight - 1) {
				neighbours = append(neighbours, Position{head.Y + 1, head.X})
			}

			// for squares next to snakes heads...
			if oLength >= mLength {
				// ...if snake is larger than us, set base to ~0
				for neighbour := range neighbours {
					yard := &neighbours[neighbour]
					matrix.Matrix[yard.X][yard.Y].Base = 0
					matrix.Matrix[yard.X][yard.Y].Danger = 1
				}
			} else if mLength > oLength {
				// ...if snake is smaller than us, set danger to -1
				for neighbour := range neighbours {
					yard := &neighbours[neighbour]
					matrix.Matrix[yard.X][yard.Y].Danger = -1
				}
			}
		}
		matrix.Matrix[head.Y][head.X].Tenure = oLength - 1

		for p := range snake.Body[1:oLength] {
			tail := &snake.Body[p]
			self := id == data.You.ID
			matrix.Matrix[tail.Y][tail.X].Tenure = oLength - 1 - p
			if self {
				matrix.Matrix[tail.Y][tail.X].Self = self
			}
		}
	}

	// limit depth by snake length
	var localDepth int = 12
	if mLength < 50 {
		if localDepth > mLength + 2 {
			localDepth = mLength + 2
		}
	} else {
		localDepth += (mLength - 30) / 18
	}

	// concurrently rate potential moves
	ch := make(chan Packet)
	defer close(ch)
	for _, direction := range directions {
		go func(direction Direction) {
			var rating = matrix.rateSquare(
				move(Position{mY, mX}, direction),
				flip(direction),
				1,
				localDepth,
				mLength,
				0,
				data.You.Health,
				[]Position{},
			)
			ch <- Packet{direction, rating}
		}(direction)	
	}

	// choose best move
	var next Direction
	var confidence float64 = 0
	var reach int = 0

	for i := 0; i < len(directions); i++ {
		packet := <-ch
		// prefer highest rated move
		if packet.Rating.Value > confidence {
			next = packet.Direction
			confidence = packet.Rating.Value
		// logic for longest path fallback if death is inevitable (ie. confidence == 0)
		} else if confidence == 0 && packet.Rating.Distance > reach {
			next = packet.Direction
			reach = packet.Rating.Distance
		}
	}

	switch next {
	case Up:
		return "up"
	case Left:
		return "left"
	case Right:
		return "right"
	case Down:
		return "down"
	default:
		return "up"
	}
}