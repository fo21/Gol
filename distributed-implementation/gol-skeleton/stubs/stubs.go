package stubs

import "uk.ac.bris.cs/gameoflife/util"

//syntax meaning ---- handler = "exported_type.exported_method"
var ProcessTurnsHandler = "GameOfLifeOperations.ProcessTurns"

type Request struct {
	InitialWorld                   [][]byte
	Turns, ImageHeight, ImageWidth int
}

type Response struct {
	FinalWorld     [][]byte
	CompletedTurns int
	AliceCells     []util.Cell
}
