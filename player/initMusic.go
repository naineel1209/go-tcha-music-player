package player

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readMusicFiles(rootDir string) []string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//get the present working directory
	finalPath := filepath.Join(pwd, rootDir)

	dirData, err := os.ReadDir(finalPath)
	if err != nil {
		panic(err)
	}

	var musicFiles []string
	var musicExtensions = []string{".mp3", ".wav", ".flac", ".ogg"}

	for _, file := range dirData {
		if file.IsDir() {
			//if it is a directory, then we need to read the files in the directory
			//and then get the music files
			musicFiles = append(musicFiles, readMusicFiles(filepath.Join(finalPath, file.Name()))...)
		} else {
			//if it is a file, then we need to check if it is a music file or not
			//if it is a music file, then we need to add it to the musicFiles list

			for _, ext := range musicExtensions {
				if strings.HasSuffix(file.Name(), ext) {
					musicFiles = append(musicFiles, file.Name())
					break //break the loop if the file is a music file
				}
			}
		}
	}

	return musicFiles
}

func InitPlayer(rootDir string) []string {

	//read the directory and get the list of music files
	musicFiles := readMusicFiles(rootDir)

	fmt.Println("The music files are: ", musicFiles)

	return musicFiles
}
