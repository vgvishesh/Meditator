package main

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	term "github.com/nsf/termbox-go"
	"log"
	"os"
	"time"
)

var isOver = false

func reset() {
	if !isOver {
		term.Sync() // cosmestic purpose
	}
}

const (
	hitCount = 108
)

func main() {
	input := os.Args[1]
	//fmt.Println("args ", input)

	f, err := os.Open(input)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	count := 0

	over := make(chan bool)
	play := make(chan bool)
	var playing = false

	err = term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()

	fmt.Println("Enter any key to see their ASCII code or press ESC button to quit")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()
		//playbackloop:
		for {
			switch ev := term.PollEvent(); ev.Type {
			case term.EventKey:
				switch ev.Key {
				case term.KeyEnter:
					play <- false
				case term.KeyEsc:
					over <- true
				case term.KeyArrowUp, term.KeyArrowDown, term.KeyArrowRight, term.KeyArrowLeft:
					count++
				default:
					fmt.Println("ASCII : ", ev.Ch)
				}
				if count < hitCount {
					reset()
					fmt.Println("Breath Count: ", count)
				} else {
					if !playing {
						reset()
						fmt.Println("Breath Count: ", count)
						play <- true
						isOver = true
					}
				}
			case term.EventError:
				panic(ev.Err)
			}
		}
	}()

	done := make(chan bool)
	var wimhoffStart time.Time
	breadingStarted := false
	go func() {
		for {
			select {
			case play := <-play:
				if play {
					reset()
					wimhoffStart = time.Now()
					playing = true
					fmt.Println("Congratulations for 108 breaths")
					fmt.Println("now stop breathing")
					speaker.Play(beep.Seq(streamer, beep.Callback(func() {
						done <- true
					})))
				} else {
					if !breadingStarted {
						stopTime := time.Now()
						println("Wim-hoff breadth stop time : ", stopTime.Sub(wimhoffStart).String())
						breadingStarted = true
					}
				}
			}
		}
	}()

	select {
	case <-done:
	case <-over:
		break
	}
}
