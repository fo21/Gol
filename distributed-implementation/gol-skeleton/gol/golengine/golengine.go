package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/gol/stubs"
)

/** add game of life functions here**/

//these are functions - can't be accessed by local controler via rpc
//this will be updateworld aka calculateNextState

/* ----------------- template function
func ReverseString(s string, i int) string {
	time.Sleep(time.Duration(rand.Intn(i)) * time.Second)
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
*/
type GameOfLifeOperations struct{}

//these are methods wow - can be accessed by local controller via rpc
func (s *GameOfLifeOperations) ProcessTurns(req stubs.Request, res *stubs.Response) (err error) {
	if req.Message == "" {
		err = errors.New("A message must be specified")
		return
	}

	fmt.Println("Got Message: " + req.Message)
	res.Message = ReverseString(req.Message, 10)
	return
}

/* -------- template method
func (s *SecretStringOperations) FastReverse(req stubs.Request, res *stubs.Response) (err error) {
	if req.Message == "" {
		err = errors.New("A message must be specified")
		return
	}

	res.Message = ReverseString(req.Message, 2)
	return
}
*/
func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	//when you register this type, its methods will be able to be called remotely
	rpc.Register(&GameOfLifeOperations{})

	//this closes it when all the things are done
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
