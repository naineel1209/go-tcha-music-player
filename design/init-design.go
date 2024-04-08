package design

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	types "github.com/naineel1209/golang-music-player/type-defs"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

var (
	globalTicker *time.Ticker
	Logger       zerolog.Logger
	loggerOnce   sync.Once
)

func init() {
	loggerOnce.Do(func() {
		file, err := os.OpenFile("logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}

		Logger = zerolog.New(file).With().Timestamp().Logger()
	})
}

func createOrGetTicker() *time.Ticker {
	if globalTicker == nil {
		globalTicker = time.NewTicker(time.Second)
	}

	return globalTicker
}

func handleTicker(base *types.BaseStruct, uiStruct *types.UiStruct) *time.Ticker {
	ticker := createOrGetTicker()

	go func() {
		for range ticker.C {
			//handle the queue
			handleQueue(base, uiStruct.Queue)
			//handle the music progress
			handleMusicProgress(base, uiStruct)

			//force the grid to redraw
			uiStruct.App.Draw() //force the grid to redraw
		}
	}()

	return ticker
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func newPrimitive(text string) *tview.TextView {
	textView := tview.NewTextView()
	textView.SetText(text)
	textView.SetTextColor(tcell.ColorYellow.TrueColor())
	textView.SetBorder(true)
	textView.SetBackgroundColor(tcell.ColorBlack.TrueColor())
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetDynamicColors(true)

	// Apply bold style to the text
	str := fmt.Sprintf("[%s]", text)
	Logger.Info().Msg(str) //log the message
	textView.SetText(str)

	return textView
}

func generateProgressBar(name string) *types.ProgressBar {
	progress := types.ProgressBar{
		Name:     name,
		Tview:    tview.NewTextView(),
		Full:     100,
		Current:  0,
		Progress: make(chan int), //unbuffered channel so that the progress bar is updated in real time
	}

	progress.Tview.SetBackgroundColor(tcell.ColorBlack.TrueColor())
	progress.Tview.SetBorder(true)
	progress.Tview.SetTitle(name)
	progress.Tview.SetTitleAlign(tview.AlignLeft)
	progress.Tview.SetDynamicColors(true)
	progress.Tview.SetTextColor(tcell.ColorYellow.TrueColor())
	progress.Tview.SetText(fmt.Sprintf("[%s] ", name))

	//channels are used to communicate between goroutines - here we are using a channel to communicate between the main goroutine and the goroutine that updates the progress bar
	go func() {
		for val := range progress.Progress {
			progress.Current = val

			progress.Tview.Clear()
			progress.Tview.SetText(fmt.Sprintf("[%s] %s%s", progress.Name, strings.Repeat("=", progress.Current), strings.Repeat("·", progress.Full-progress.Current)))
		}
	}()

	return &progress
}

func generateMusicProgress() (*tview.Grid, []*tview.TextView, *types.ProgressBar) {
	musicProgress := tview.NewGrid()

	// Set background color and text color for individual list items
	musicProgress.SetBackgroundColor(tcell.ColorBlack.TrueColor())
	musicProgress.SetBorder(true)
	musicProgress.SetTitle("Progress - Bar")

	// Create a new text view
	currTime := newPrimitive("-- : --")  // Create a new text view
	totalTime := newPrimitive("-- : --") // Create a new text view
	progressBar := generateProgressBar("Music Progress")

	// Add the text view to the grid
	//TODO: add the progress bar
	musicProgress.AddItem(currTime, 0, 0, 1, 3, 0, 0, false)
	musicProgress.AddItem(progressBar.Tview, 0, 3, 1, 24, 0, 0, false)
	musicProgress.AddItem(totalTime, 0, 27, 1, 3, 0, 0, false)

	return musicProgress, []*tview.TextView{currTime, totalTime, progressBar.Tview}, progressBar
}

func handleMusicProgress(base *types.BaseStruct, uiStruct *types.UiStruct) {
	//handle the music progress
	//1. get the queue
	q := base.Q

	if (q.GetCurrentStreamer()) == nil {

		Logger.Info().Msg("No music is playing") //log the message

		uiStruct.MusicProgress[0].SetText("[-- : --]")
		uiStruct.MusicProgress[1].SetText("[-- : --]")
		uiStruct.ProgressBar.Name = "[Music Progress]" //set the name of the progress bar
		uiStruct.ProgressBar.Progress <- 0             //send the progress to the progress bar	//send the progress to the progress bar

		return
	}

	//2. get the current streamer and type assert it to beep.StreamSeekCloser
	currStreamer := (q.GetCurrentStreamer()).(beep.StreamSeekCloser)
	currName := q.GetCurrentName()

	//4. get the current time and total time
	format := q.GetCurrentFormat()                                                    //get the format of the current streamer
	currTime := format.SampleRate.D(currStreamer.Position()).Round(time.Second / 100) //get the current time of the music Rounding to 1/100th of a second
	totalTime := format.SampleRate.D(currStreamer.Len()).Round(time.Second / 100)     //get the total time of the music Rounding to 1/100th of a second

	currTimeSeconds := int(math.Floor(currTime.Seconds())) % 60
	currTimeMinutes := int(math.Floor(currTime.Minutes())) % 60
	totalTimeSeconds := int(math.Floor(totalTime.Seconds())) % 60
	totalTimeMinutes := int(math.Floor(totalTime.Minutes())) % 60

	//5. update the music progress
	uiStruct.MusicProgress[0].SetText(fmt.Sprintf("[%vm:%vs]", currTimeMinutes, currTimeSeconds))
	uiStruct.MusicProgress[1].SetText(fmt.Sprintf("[%vm:%vs]", totalTimeMinutes, totalTimeSeconds))

	//handle the progress bar
	progress := uiStruct.ProgressBar
	progress.Name = currName
	progress.Progress <- int(math.Floor((currTime.Seconds() / totalTime.Seconds()) * 100)) //send the progress to the progress bar
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

	// playlist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	switch event.Key() {
	// 	case tcell.KeyTAB:

	// 	default: // do nothing
	// 	}

	// 	return event
	// })

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
		//ctrl and resampled are completed in the Add method of the queue

		//3. add the streamer to the queue

		speaker.Lock()
		base.Q.Add(streamer, format, mainText)
		speaker.Unlock()

		//5. handle the queue
		handleQueue(base, uiStruct.Queue)
		handleMusicProgress(base, uiStruct) //handle the music progress
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
	musicProgressGrid, musicProgress, progressBar := generateMusicProgress()
	uiStruct := &types.UiStruct{
		Queue:         generateQueue(),
		Playlist:      generatePlaylist(base.Paths),
		MusicProgress: musicProgress,
		App:           app,
		ProgressBar:   progressBar,
	}

	//place the items into their place
	//set the position for the items
	grid.AddItem(uiStruct.Playlist, 0, 0, 3, 10, 0, 0, false)
	grid.AddItem(uiStruct.Queue, 0, 10, 3, 20, 0, 0, false)
	grid.AddItem(musicProgressGrid, 3, 0, 1, 30, 0, 0, false)

	handleMusic(base, uiStruct)            //register the music handler
	ticker := handleTicker(base, uiStruct) //handle the ticker
	defer ticker.Stop()                    //stop the ticker when the application stops

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
