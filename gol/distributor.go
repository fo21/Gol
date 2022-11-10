package gol

import (
	"strconv"
	"sync"
	"uk.ac.bris.cs/gameoflife/util"

	"time"
)

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

//original version of calculateNextState()
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

func calculateNextStateByThread(p Params, world [][]byte, startY, endY int) [][]byte {

	newWorld := make([][]byte, p.ImageWidth)
	for i := range newWorld {
		newWorld[i] = make([]byte, endY+1)
	}

	for i := startY; i < endY; i++ {
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

func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	newCell := []util.Cell{}
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			if world[i][j] == 0xff {
				newCell = append(newCell, util.Cell{j, i})
			}
		}
	}
	return newCell
}

func calculateCount(p Params, world [][]byte) int {
	sum := 0
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			if world[i][j] != 0 {
				sum++
			}
		}
	}
	return sum
}

func worker(p Params, world [][]byte, startX, endX, startY, endY int, out chan<- [][]uint8) {
	imagePart := calculateNextStateByThread(p, world, startY, endY)
	out <- imagePart
}

func makeMatrix(height, width int) [][]uint8 {
	matrix := make([][]uint8, height)
	for i := range matrix {
		matrix[i] = make([]uint8, width)
	}
	return matrix
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.

	imageHeight := p.ImageHeight
	imageWidth := p.ImageWidth

	heightString := strconv.Itoa(imageHeight)
	widthString := strconv.Itoa(imageWidth)

	filename := heightString + "x" + widthString

	c.ioCommand <- ioInput

	c.ioFilename <- filename

	world := make([][]uint8, imageHeight)
	for i := 0; i < imageHeight; i++ {
		world[i] = make([]uint8, imageWidth)
		for j := range world[i] {
			byte := <-c.ioInput
			world[i][j] = byte
		}
	}

	turns := p.Turns
	turn := 0

	// TODO: Execute all turns of the Game of Life.

	var m sync.Mutex

	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				m.Lock()
				c.events <- AliveCellsCount{turn, calculateCount(p, world)}
				m.Unlock()
			}
		}
	}()

	if p.Threads == 1 {

		for turn < turns {
			m.Lock()
			world = calculateNextState(p, world)
			turn++
			c.events <- TurnComplete{CompletedTurns: turn}
			m.Unlock()
		}
	} else {
		for turn < turns {
			workerHeight := p.ImageHeight / p.Threads
			out := make([]chan [][]uint8, p.Threads)
			for i := range out {
				out[i] = make(chan [][]uint8)
			}

			m.Lock()
			for i := 0; i < p.Threads; i++ {
				go worker(p, world, i*workerHeight, (i+1)*workerHeight, 0, p.ImageWidth, out[i])
			}

			var newPixelData [][]uint8

			newPixelData = makeMatrix(0, 0)

			for i := 0; i < p.Threads; i++ {
				part := <-out[i]
				newPixelData = append(newPixelData, part...)
			}
			world = newPixelData
			turn++
			c.events <- TurnComplete{CompletedTurns: turn}
			m.Unlock()
		}

	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

	alive := calculateAliveCells(p, world)
	c.events <- FinalTurnComplete{turns, alive}
	ticker.Stop()
	done <- true

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
