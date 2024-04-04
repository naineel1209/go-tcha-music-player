package main

import (
	"fmt"

	"github.com/naineel1209/golang-music-player/design"
	"github.com/naineel1209/golang-music-player/player"
)

func main() {
	//root directory
	rootDir := "/music"

	//read the directory and get the list of music files
	musicFiles := player.InitPlayer(rootDir)

	//inited the design - it should return us the grid object of tview lib
	grid := design.InitDesign(musicFiles)

	fmt.Printf("the grid object is: %v\n", grid)
}
