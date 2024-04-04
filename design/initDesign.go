package design

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func newPrimitive(text string) tview.Primitive {
	textView := tview.NewTextView()
	textView.SetText(text)
	textView.SetTextColor(tcell.ColorYellow.TrueColor())
	textView.SetBackgroundColor(tcell.ColorBlack.TrueColor())
	textView.SetBorder(true)
	textView.SetTitle(text)

	return textView
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

func handleMusic(musicProgress *tview.TextView, playlist *tview.List) {
	playlist.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		musicProgress.SetText(mainText)
	})
}

func InitDesign(musicFiles []string) *tview.Grid {
	//new grid
	app := tview.NewApplication()

	grid := tview.NewGrid()

	//items that can be added to grid
	playlist := generatePlaylist(musicFiles)
	queue := newPrimitive("Queue")
	musicProgress := newPrimitive("Music Progress")

	//place the items into their place
	//set the position for the items
	grid.AddItem(playlist, 0, 0, 3, 10, 0, 0, false)
	grid.AddItem(queue, 0, 10, 3, 20, 0, 0, false)
	grid.AddItem(musicProgress, 3, 0, 1, 30, 0, 0, false)

	handleMusic((musicProgress).(*tview.TextView), playlist) //custom interface casting - need to call only once

	items := []tview.Primitive{playlist, queue, musicProgress}
	i := 0
	app.SetFocus(items[i])

	// how to enable the arrow keys to navigate the grid
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			app.SetFocus(items[i])
		case tcell.KeyCtrlC:
			app.Stop()
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
