package gol

import (
	"fmt"
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
	keyPresses <-chan rune
}

//Counts the alive neighbours of a cell
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

//check if cell is alive
func isAlive(cell byte) bool {
	if cell == 255 {
		return true
	}
	return false
}

//calculates the next state of the world for multiple threaded
func calculateNextState(p Params, world [][]byte, startY, endY int) [][]byte {

	newWorld := make([][]byte, p.ImageWidth)
	for x := range newWorld {
		newWorld[x] = make([]byte, endY-startY) // calculate height of this part of matrix
	}

	if endY > p.ImageHeight {
		endY = p.ImageHeight
	}

	for x := 0; x < p.ImageWidth; x++ {
		for y := 0; y < endY-startY; y++ {
			//fmt.Println("=========\nendY: " + strconv.Itoa(endY) + " startY: " + strconv.Itoa(startY))
			//fmt.Println("y is currently: " + strconv.Itoa(y))
			neighbours := getLiveNeighbours(p, world, x, y+startY)
			//fmt.Println("Working on Cell with X: " + strconv.Itoa(x) + " and Y: " + strconv.Itoa(y))
			if world[x][y+startY] == 0xff && (neighbours < 2 || neighbours > 3) {
				newWorld[x][y] = 0x0
			} else if world[x][y+startY] == 0x0 && neighbours == 3 {
				newWorld[x][y] = 0xff
			} else {
				newWorld[x][y] = world[x][y+startY]
			}
		}
	}
	return newWorld
}

//create a list of alive cells
func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	var newCell []util.Cell
	for x := 0; x < p.ImageWidth; x++ {
		for y := 0; y < p.ImageHeight; y++ {
			if world[x][y] == 0xff {
				newCell = append(newCell, util.Cell{X: y, Y: x}) // TODO FIGURE THIS OUT
			}
		}
	}
	return newCell
}

//count the number of alive cells
func calculateCount(p Params, world [][]byte) int {
	sum := 0
	for x := 0; x < p.ImageWidth; x++ {
		for y := 0; y < p.ImageHeight; y++ {
			if world[x][y] != 0 {
				sum++
			}
		}
	}
	return sum
}

//worker function for multiple threaded case
func worker(p Params, world [][]byte, startY, endY int, out chan<- [][]uint8) {
	imagePart := calculateNextState(p, world, startY, endY)
	out <- imagePart
}

//report the CellFlipped event when a cell changes state
func compareWorlds(old, new [][]byte, c *distributorChannels, turn int, p Params) {
	for x := 0; x < p.ImageWidth; x++ {
		//fmt.Printf("x: " + strconv.Itoa(x))
		for y := 0; y < p.ImageHeight; y++ {
			//fmt.Printf(" y: " + strconv.Itoa(y) + "\n")
			if old[x][y] != new[x][y] {
				c.events <- CellFlipped{turn, util.Cell{X: y, Y: x}}
			}
		}
	}
}

func saveWorld(p Params, c distributorChannels, world [][]byte, turns int) {
	c.ioCommand <- ioOutput
	c.ioFilename <- strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(turns)

	for x := 0; x < p.ImageWidth; x++ {
		for y := 0; y < p.ImageHeight; y++ {
			c.ioOutput <- world[x][y]
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	imageHeight := p.ImageHeight
	imageWidth := p.ImageWidth

	c.ioCommand <- ioInput

	var filename = strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioFilename <- filename

	world := make([][]byte, imageWidth)
	for x := 0; x < imageWidth; x++ {
		world[x] = make([]byte, imageHeight)
		for y := 0; y < p.ImageHeight; y++ {
			byte := <-c.ioInput
			world[x][y] = byte
		}
	}

	turns := p.Turns
	turn := 0

	for _, cell := range calculateAliveCells(p, world) {
		c.events <- CellFlipped{0, cell}
	}

	var m sync.Mutex

	//ticker that reports the number of alive cells every two seconds

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

	//keypress
	var ok = 1
	go func() {
		for {
			switch <-c.keyPresses {
			case 'p':
				if ok == 1 {
					fmt.Printf("The current turn that is being processed: %d\n", turn)
					ok = 0
					m.Lock()
				} else if ok == 0 {
					fmt.Println("Continuing... \n")
					ok = 1
					m.Unlock()
				}
			case 'q':
				if ok == 1 {
					saveWorld(p, c, world, turn)

					alive := calculateAliveCells(p, world)
					c.events <- FinalTurnComplete{turn, alive}

					ticker.Stop()
					done <- true
					// Make sure that the Io has finished any output before exiting.
					c.ioCommand <- ioCheckIdle
					<-c.ioIdle

					c.events <- StateChange{turn, Quitting}

					// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
					close(c.events)
				} else {
					fmt.Println("Pressed the wrong key. try again \n")
				}
			case 's':
				if ok == 1 {
					saveWorld(p, c, world, turn)
				} else {
					fmt.Println("Pressed wrong key. try again \n")
				}
			}
		}
	}()

	//calculate next state depending on the number of threads

	if p.Threads == 1 {

		for turn < turns {
			m.Lock()
			oldWorld := append(world)
			world = calculateNextState(p, world, 0, p.ImageHeight)
			compareWorlds(oldWorld, world, &c, turn+1, p)
			turn++
			c.events <- TurnComplete{CompletedTurns: turn}
			m.Unlock()
		}
	} else {

		workerHeight := p.ImageHeight / p.Threads
		if p.ImageHeight%p.Threads > 0 {
			workerHeight++
		}
		//fmt.Println("Starting image: " + filename + " with " + strconv.Itoa(p.Threads) + " threads and a worker height of: " + strconv.Itoa(workerHeight))

		out := make([]chan [][]uint8, p.Threads)
		for i := range out {
			out[i] = make(chan [][]uint8)
		}

		for turn < turns {
			m.Lock()

			for i := 0; i < p.Threads; i++ {
				//fmt.Println("Starting worker between Y: " + strconv.Itoa(i*workerHeight) + ", " + strconv.Itoa((i+1)*workerHeight))
				go worker(p, world, i*workerHeight, (i+1)*workerHeight, out[i])
			}

			var newPixelData [][]uint8

			newPixelData = make([][]byte, imageWidth)
			for x := 0; x < imageWidth; x++ {
				newPixelData[x] = make([]byte, imageHeight)
			}

			for i := 0; i < p.Threads; i++ {
				var yOffset = i * workerHeight
				part := <-out[i]

				for x := 0; x < p.ImageWidth; x++ {
					//fmt.Printf("x: " + strconv.Itoa(x) + "\n")
					for y := 0; y < workerHeight; y++ {
						var yAdjusted = yOffset + y
						if yAdjusted < imageHeight {
							//fmt.Printf(" y: " + strconv.Itoa(y))
							newPixelData[x][y+yOffset] = part[x][y]
						}
					}
				}
			}

			compareWorlds(world, newPixelData, &c, turn, p)
			world = newPixelData
			turn++
			c.events <- TurnComplete{CompletedTurns: turn}
			m.Unlock()

		}

	}

	c.events <- FinalTurnComplete{turn, calculateAliveCells(p, world)}

	ticker.Stop()
	done <- true

	//write final state of the world to pgm image

	saveWorld(p, c, world, turns)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
