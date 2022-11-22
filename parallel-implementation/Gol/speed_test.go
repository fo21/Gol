package main

import (
	"fmt"
	"strconv"
	"testing"
	"uk.ac.bris.cs/gameoflife/gol"
)

// Benchmark applies the filter to the ship.png b.N times.
// The time taken is carefully measured by go.
// The b.N  repetition is needed because benchmark results are not always constant.
func BenchmarkGol(b *testing.B) {
	// Disable all program output apart from benchmark results
	//os.Stdout = nil

	p := gol.Params{ImageWidth: 512, ImageHeight: 512, Turns: 1000}

	// Use a for-loop to run 5 sub-benchmarks, with 1, 2, 4, 8 and 16 workers.
	for threads := 1; threads <= 16; threads *= 2 {

		p.Threads = threads
		benchMarkName := fmt.Sprintf("speed-%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)

		b.Run(benchMarkName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				events := make(chan gol.Event)
				go gol.Run(p, events, nil)
				for event := range events {
					switch e := event.(type) {
					case gol.FinalTurnComplete:
						fmt.Println(strconv.Itoa(len(e.Alive)))
					}
				}
			}
		})
	}
}
