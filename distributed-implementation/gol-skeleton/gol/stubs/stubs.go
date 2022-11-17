package stubs

var ProcessTurnsHandler = "GameOfLifeOperations.ProcessTurns"

type Request struct {
	initialWorld                   [][]byte
	turns, imageHeight, imageWidth int
}

type Response struct {
	finalWorld [][]byte
}
