package design

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	types "github.com/naineel1209/golang-music-player/type-defs"
	"github.com/rivo/tview"
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func newPrimitive(text string) tview.Primitive {
	textView := tview.NewTextView()
	textView.SetText(text)
	textView.SetTextColor(tcell.ColorYellow.TrueColor())
	textView.SetBackgroundColor(tcell.ColorBlack.TrueColor())
	textView.SetBorder(true)
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetDynamicColors(true)

	return textView
}

func generateMusicProgress() (*tview.Grid, []*tview.Primitive) {
	musicProgress := tview.NewGrid()

	// Set background color and text color for individual list items
	musicProgress.SetBackgroundColor(tcell.ColorBlack.TrueColor())
	musicProgress.SetBorder(true)
	musicProgress.SetTitle("Progress - Bar")

	// Create a new text view
	currTime := newPrimitive("-- : --")     // Create a new text view
	totalTime := newPrimitive("-- : --")    // Create a new text view
	musicName := newPrimitive("Music Name") // Create a new text view

	// Add the text view to the grid
	musicProgress.AddItem(currTime, 0, 0, 1, 10, 0, 0, false)
	musicProgress.AddItem(musicName, 0, 10, 1, 10, 0, 0, false)
	musicProgress.AddItem(totalTime, 0, 20, 1, 10, 0, 0, false)

	return musicProgress, []*tview.Primitive{&currTime, &totalTime, &musicName}
}

func handleMusicProgress(base *types.BaseStruct, uiStruct *types.UiStruct) {
	//handle the music progress
	//1. get the queue
	q := base.Q

	//2. get the current streamer and type assert it to beep.StreamSeekCloser
	currStreamer := (q.GetCurrentStreamer()).(beep.StreamSeekCloser)
	currName := q.GetCurrentName()

	//4. get the current time and total time
	format := q.GetCurrentFormat()                                                    //get the format of the current streamer
	currTime := format.SampleRate.D(currStreamer.Position()).Round(time.Second / 100) //get the current time of the music
	totalTime := format.SampleRate.D(currStreamer.Len()).Round(time.Second / 100)     //get the total time of the music

	//5. update the music progress
	(*uiStruct.MusicProgress[0]).(*tview.TextView).SetText(currTime.String())
	(*uiStruct.MusicProgress[1]).(*tview.TextView).SetText(totalTime.String())
	(*uiStruct.MusicProgress[2]).(*tview.TextView).SetText(currName)
}

func generatePlaylist(musicFiles []string) *tview.List {
	playlist := tview.NewList()

	// Set background color and text color for individual list items
	playlist.SetBackgroundColor(tcell.ColorBlack.TrueColor())
	playlist.SetMainTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite.TrueColor()))
	playlist.SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorYellow.TrueColor()).Bold(true).Underline(true))
	playlist.SetShortcutStyle(tcell.StyleDefault.Foreground(tcell.ColorGreen.TrueColor()))

	for idx, file := range musicFiles {
		playlist.AddItem(file, "", rune(49+idx), nil)
	}

	playlist.SetTitle("Playlist")
	playlist.SetTitleAlign(tview.AlignLeft)
	playlist.SetBorder(true)

	playlist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:

		default: // do nothing
		}

		return event
	})

	return playlist
}

func generateQueue() *tview.List {
	queue := tview.NewList()

	// Set background color and text color for individual list items
	queue.SetBackgroundColor(tcell.ColorBlack.TrueColor())
	queue.SetMainTextStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite.TrueColor()))
	queue.SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorYellow.TrueColor()).Bold(true).Underline(true))
	queue.SetShortcutStyle(tcell.StyleDefault.Foreground(tcell.ColorGreen.TrueColor()))

	queue.SetTitle("Queue")
	queue.SetTitleAlign(tview.AlignCenter)
	queue.SetBorder(true)

	return queue
}

func handleQueue(base *types.BaseStruct, queue *tview.List) {
	//1. get the queue
	q := base.Q

	//2. clear the queue
	queue.Clear()

	//traverse the queue and add the items to the queue
	for idx := range q.Streamers {
		queue.AddItem(q.Streamers[idx].Name, "", rune(49+idx), nil)
	}

	//auto focus on the first item
	queue.SetCurrentItem(0)
}

func handleMusic(base *types.BaseStruct, uiStruct *types.UiStruct) {
	uiStruct.Playlist.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		//play the music -

		//1. get the file path using the index and read the file
		fi, ext, err := getFilePath(base, index)
		handleError(err)

		var streamer beep.StreamSeekCloser //streamer streams the audio data from the file as and when required
		var format beep.Format             //format contains the audio data and sample rate - basically the audio data format

		//2. create a streamer
		switch ext {
		case ".mp3":
			//decode mp3
			streamer, format, err = mp3.Decode(fi)
			handleError(err)
		case ".wav":
			//decode wav
			streamer, format, err = wav.Decode(fi)
			handleError(err)
		case ".flac":
			//decode flac
			streamer, format, err = flac.Decode(fi)
			handleError(err)
		default:
			panic("Unsupported file format")
		}

		//make the streamer a ctrl and volume
		ctrl := &beep.Ctrl{Streamer: streamer, Paused: false} //control the streamer
		// volume := &effects.Volume{Streamer: ctrl, Base: 2, Volume: 0}   //volume control
		// resampled := beep.Resample(4, format.SampleRate, 44100, volume) //resample the audio data to 44100 Hz (standard sample rate)

		//3. add the streamer to the queue

		speaker.Lock()
		base.Q.Add(ctrl, format, mainText)
		speaker.Unlock()

		//5. handle the queue
		handleQueue(base, uiStruct.Queue)

		// TODO: Solve the error here   //7. handle the music progress
		// go func() {
		// 	for {
		// 		handleMusicProgress(base, uiStruct)
		// 		time.Sleep(1 * time.Second)
		// 	}
		// }()
	})
}

func getFilePath(base *types.BaseStruct, index int) (*os.File, string, error) {
	filePath := base.Paths[index]
	rootDir, err := os.Getwd()
	handleError(err)
	filePath = filepath.Join(rootDir, "/music", filePath)
	fi, err := os.Open(filePath)
	handleError(err)

	ext := filepath.Ext(fi.Name())

	return fi, ext, nil
}

func InitDesign(base *types.BaseStruct) *tview.Grid {
	//new grid
	app := tview.NewApplication()

	grid := tview.NewGrid()

	//items that can be added to grid
	musicProgressGrid, musicProgress := generateMusicProgress()
	uiStruct := &types.UiStruct{
		Queue:         generateQueue(),
		Playlist:      generatePlaylist(base.Paths),
		MusicProgress: musicProgress,
	}

	//place the items into their place
	//set the position for the items
	grid.AddItem(uiStruct.Playlist, 0, 0, 3, 10, 0, 0, false)
	grid.AddItem(uiStruct.Queue, 0, 10, 3, 20, 0, 0, false)
	grid.AddItem(musicProgressGrid, 3, 0, 1, 30, 0, 0, false)

	handleMusic(base, uiStruct) //register the music handler

	items := []tview.Primitive{musicProgressGrid, uiStruct.Playlist, uiStruct.Queue}
	i := len(items) - 1

	// how to enable the arrow keys to navigate the grid
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			app.SetFocus(items[i])
		case tcell.KeyCtrlC:
			app.Stop()
			return event
		}

		i = (i + 1) % len(items)

		return event
	})

	//apply the boxy methods
	grid.SetBackgroundColor(tcell.ColorBlack.TrueColor()).SetBorderAttributes(tcell.AttrItalic | tcell.AttrBlink)

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

	return grid
}
