package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"codeberg.org/anaseto/goal"
	gos "codeberg.org/anaseto/goal/os"
	"github.com/eiannone/keyboard"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//go:embed k.goal
var kGoalContents string

func main() {
	var gameFile string
	flag.StringVar(&gameFile, "game", "", "Game file to load")
	flag.StringVar(&gameFile, "g", "", "Game file to load (shorthand)")
	flag.Parse()

	if gameFile == "" {
		fmt.Println("Please provide a file name with --game or -g")
		flag.Usage()
		os.Exit(1)
	}

	contents, err := os.ReadFile(gameFile)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", gameFile, err)
		os.Exit(1)
	}
	gameGoalSource := string(contents)

	fmt.Printf("Loading game file: %s\n", gameFile)
	ctx := goal.NewContext()
	ctx.Log = os.Stderr
	gos.Import(ctx, "")

	_, err = ctx.EvalPackage(kGoalContents, "<builtin>", "k")
	if err != nil {
		fmt.Printf("Error loading k.goal: %v\n", err)
	}

	_, err = ctx.Eval(gameGoalSource)
	if err != nil {
		fmt.Printf("Error evaluating Goal game source from file %s: %v\n", gameFile, err)
	}

	if err := keyboard.Open(); err != nil {
		fmt.Println("Failed to open keyboard:", err)
		return
	}
	defer keyboard.Close()

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	keyEvents, err := keyboard.GetKeys(10)
	if err != nil {
		fmt.Println("Failed to get keys:", err)
		return
	}

	_, err = ctx.Eval("render[board]")
	if err != nil {
		fmt.Printf("Error rendering initial board: %v\n", err)
	}

	// Open the audio file
	file, err := os.Open("cat.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Decode the audio file
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	// Initialize the speaker
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	// Play the audio
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	// Wait for audio to finish playing
	// <-done

	for {
		select {
		case <-tick.C:
		case event := <-keyEvents:
			if event.Err != nil {
				fmt.Println("Keyboard error:", event.Err)
				return
			}
			if event.Rune == 'q' {
				fmt.Println("Bye!")
				os.Exit(0)
			}
			switch event.Key {
			case keyboard.KeyArrowUp:
				_, err = ctx.Eval(update("n"))
				if err != nil {
					fmt.Printf("Error moving n: %v\n", err)
				}
			case keyboard.KeyArrowDown:
				_, err = ctx.Eval(update("s"))
				if err != nil {
					fmt.Printf("Error moving s: %v\n", err)
				}
			case keyboard.KeyArrowLeft:
				_, err = ctx.Eval(update("w"))
				if err != nil {
					fmt.Printf("Error moving w: %v\n", err)
				}
			case keyboard.KeyArrowRight:
				_, err = ctx.Eval(update("e"))
				if err != nil {
					fmt.Printf("Error moving e: %v\n", err)
				}
			case keyboard.KeyEnter:
				_, err = ctx.Eval(`reset""; render[board]`)
				if err != nil {
					fmt.Printf("Error resetting board: %v\n", err)
				}
			case keyboard.KeyEsc:
				fmt.Println("Later!")
				os.Exit(0)
			}
		}
	}
}

func update(dir string) string {
	return fmt.Sprintf("update[board;\"%s\"]; render[board]", dir)
}
