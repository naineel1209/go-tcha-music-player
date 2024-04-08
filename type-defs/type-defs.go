package types

import (
	"github.com/gopxl/beep"
	"github.com/rivo/tview"
)

type BaseStruct struct {
	Paths []string
	Q     *Queue
}

type UiStruct struct {
	Queue         *tview.List
	Playlist      *tview.List
	MusicProgress []*tview.TextView
	App           *tview.Application
	ProgressBar   *ProgressBar
}

type ProgressBar struct {
	Name     string
	Tview    *tview.TextView
	Full     int
	Current  int
	Progress chan int
}

type CustomStreamer struct {
	Streamer  beep.Streamer
	Resampled *beep.Resampler
	Ctrl      *beep.Ctrl
	Format    beep.Format
	Name      string
}

type Queue struct {
	Streamers []CustomStreamer
}

func (q *Queue) Add(streamer beep.Streamer, format beep.Format, name string) {
	q.Streamers = append(q.Streamers, CustomStreamer{
		Streamer:  streamer,
		Resampled: beep.Resample(4, format.SampleRate, 44100, streamer), //resample the audio data to 44100 Hz (standard sample rate
		Ctrl:      &beep.Ctrl{Streamer: streamer, Paused: false},
		Format:    format,
		Name:      name,
	})
}

func (q *Queue) Stream(samples [][2]float64) (n int, ok bool) {
	//handle the sampling and streaming

	filled := 0
	for filled < len(samples) {
		//if there are no streamers - make a silence
		if len(q.Streamers) == 0 {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}

		//now need to stream the current streamer
		n, ok := q.Streamers[0].Resampled.Stream(samples[filled:]) //stream the samples from the current streamer
		if !ok {                                                   // not ok means the streamer has finished
			q.Streamers = q.Streamers[1:]
		}

		filled += n // adding the number of samples consumed into filled
	}

	return len(samples), true
}

func (q *Queue) Next() {
	if len(q.Streamers) == 0 {
		return
	}

	//remove the current streamer but the filled value will be removed in the next iteration
	q.Streamers = q.Streamers[1:]
}

func (q *Queue) Err() error {
	panic("implement me")
}

func (q *Queue) Len() int {
	return len(q.Streamers)
}

func (q *Queue) GetCurrentStreamer() beep.Streamer {
	if len(q.Streamers) == 0 {
		return nil
	}

	return q.Streamers[0].Streamer
}

func (q *Queue) GetCurrentName() string {
	if len(q.Streamers) == 0 {
		return ""
	}

	return q.Streamers[0].Name
}

func (q *Queue) GetCurrentFormat() beep.Format {
	if len(q.Streamers) == 0 {
		return beep.Format{}
	}

	return q.Streamers[0].Format
}
