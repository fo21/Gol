package gol

import (
	"log"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func makeCall(client *rpc.Client, world [][]byte) {
	request := stubs.Request{InitialWorld: world}

	//this is a pointer
	response := new(stubs.Response)
	client.Call(stubs.ProcessTurnsHandler, request, response)
	//fmt.Println("Responded: " + response.Message)
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	board := loadBoard(p, c)
	board1 := allocateBoard(p.ImageHeight, p.ImageWidth)
	turn := 0

	server := "127.0.0.1:8030"
	client, err := rpc.Dial("tcp", server)
	if err != nil {
		log.Fatal("dialing: ", err)
	}
	defer client.Close()

	/* ----------template code ----------------
	file, _ := os.Open("wordlist")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := scanner.Text()
		fmt.Println("Called: " + t)
		makeCall(client, t)
	}

	*/

	// TODO: Execute all turns of the Game of Life.

	//this goes on the gol engine along with all its functions
	for turn < p.Turns {
		updateBoard(c, turn, board, board1, p.ImageHeight, p.ImageWidth, 0, p.ImageHeight)

		turn = turn + 1
	}
	//end of what needs to be deported

	// TODO: Report the final state using FinalTurnCompleteEvent.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
