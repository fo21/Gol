package gol

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func getLiveNeighbours(p Params, world [][]byte, a, b int) int {
	var alive = 0
	var widthLeft int
	var widthRight int
	var heightUp int
	var heightDown int

	if a == 0 {
		widthLeft = p.ImageWidth - 1
	} else {
		widthLeft = a - 1
	}
	if a == p.ImageWidth-1 {
		widthRight = 0
	} else {
		widthRight = a + 1
	}

	if b == 0 {
		heightUp = p.ImageHeight - 1
	} else {
		heightUp = b - 1
	}

	if b == p.ImageHeight-1 {
		heightDown = 0
	} else {
		heightDown = b + 1
	}

	if isAlive(world[widthLeft][b]) {
		alive = alive + 1
	}
	if isAlive(world[widthRight][b]) {
		alive = alive + 1
	}
	if isAlive(world[widthLeft][heightUp]) {
		alive = alive + 1
	}
	if isAlive(world[a][heightUp]) {
		alive = alive + 1
	}
	if isAlive(world[widthRight][heightUp]) {
		alive = alive + 1
	}
	if isAlive(world[widthLeft][heightDown]) {
		alive = alive + 1
	}
	if isAlive(world[a][heightDown]) {
		alive = alive + 1
	}
	if isAlive(world[widthRight][heightDown]) {
		alive = alive + 1
	}
	return alive
}

func isAlive(cell byte) bool {
	if cell == 255 {
		return true
	}
	return false
}

func calculateNextState(p Params, world [][]byte) [][]byte {
	newWorld := make([][]byte, p.ImageWidth)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageHeight)
	}
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			neighbours := getLiveNeighbours(p, world, i, j)
			if world[i][j] == 0xff && (neighbours < 2 || neighbours > 3) {
				newWorld[i][j] = 0x0
			} else if world[i][j] == 0x0 && neighbours == 3 {
				newWorld[i][j] = 0xff
			} else {
				newWorld[i][j] = world[i][j]
			}
		}
	}
	return newWorld
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.

	imageHeight := p.ImageHeight
	imageWidth := p.ImageWidth

	world := make([][]uint8, imageHeight)
	for i := 0; i < imageHeight; i++ {
		world[i] = make([]uint8, imageWidth)
	}

	turns := p.Turns
	turn := 0

	// TODO: Execute all turns of the Game of Life.

	for turn < turns {

		turn++
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
