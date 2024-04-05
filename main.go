package main

import (
	"fmt"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/naineel1209/golang-music-player/design"
	"github.com/naineel1209/golang-music-player/player"
	types "github.com/naineel1209/golang-music-player/type-defs"
)

func createQueuePlayer(q *types.Queue) *types.Queue {
	sampleRate := beep.SampleRate(44100)
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	speaker.Play(q)
	return q
}

func main() {
	//root directory
	rootDir := "/music"

	//read the directory and get the list of music files
	musicFiles := player.InitPlayer(rootDir)

	var queue types.Queue
	q := createQueuePlayer(&queue)

	//inited the design - it should return us the grid object of tview lib
	grid := design.InitDesign(&types.BaseStruct{Paths: musicFiles, Q: q})

	fmt.Printf("the grid object is: %v\n", grid)
}
