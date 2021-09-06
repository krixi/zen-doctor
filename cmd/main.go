package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jroimartin/gocui"
	zen_doctor "github.com/krixi/zen-doctor/internal"
	"github.com/pkg/errors"
)

var (
	done = make(chan bool)
)

const (
	threatView = "threat"
)

func main() {
	rand.Seed(time.Now().Unix())
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	state := zen_doctor.NewGameState(zen_doctor.Tutorial)

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SetManagerFunc(layout(&state))

	// global ket to quit
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	// game-specific keybinds
	if err := gameKeybinds(g, &state); err != nil {
		log.Panicln(err)
	}
	// start the game loop
	go gameLoop(g, &state)

	// start the terminal display loop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(state *zen_doctor.GameState) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		level := state.GetLevel()
		x1, y1, x2, y2 := zen_doctor.CalculateViewPosition(level.Width, level.Height, maxX, maxY)

		if v, err := g.SetView(level.Name(), x1, y1, x2, y2); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Wrapf(err, "setting view for level %s", level.Name())
			}
			v.Title = level.Name()
			fmt.Fprintf(v, "%s", state.String())
			g.SetCurrentView(level.Name())
		}
		if v, err := g.SetView(threatView, x1, y1-3, x2, y1-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Threat"
		}

		//if v, err := g.SetView("colors", 0, 0, maxX-1, maxY-1); err != nil {
		//	if err != gocui.ErrUnknownView {
		//		return err
		//	}
		//	// 256-colors escape codes
		//	for i := 0; i < 256; i++ {
		//		str := fmt.Sprintf("\x1b[48;5;%dm\x1b[30m%3d\x1b[0m ", i, i)
		//		str += fmt.Sprintf("\x1b[38;5;%dm%3d\x1b[0m ", i, i)
		//
		//		if (i+1)%10 == 0 {
		//			str += "\n"
		//		}
		//
		//		fmt.Fprint(v, str)
		//	}
		//
		//	fmt.Fprint(v, "\n\n")
		//}
		return nil
	}
}

func gameKeybinds(g *gocui.Gui, state *zen_doctor.GameState) error {
	level := state.GetLevel()
	// in-game keybinds
	if err := g.SetKeybinding(level.Name(), gocui.KeyArrowUp, gocui.ModNone, movePlayer(state, zen_doctor.MoveUp)); err != nil {
		return err
	}
	if err := g.SetKeybinding(level.Name(), gocui.KeyArrowDown, gocui.ModNone, movePlayer(state, zen_doctor.MoveDown)); err != nil {
		return err
	}
	if err := g.SetKeybinding(level.Name(), gocui.KeyArrowLeft, gocui.ModNone, movePlayer(state, zen_doctor.MoveLeft)); err != nil {
		return err
	}
	if err := g.SetKeybinding(level.Name(), gocui.KeyArrowRight, gocui.ModNone, movePlayer(state, zen_doctor.MoveRight)); err != nil {
		return err
	}
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}

func movePlayer(state *zen_doctor.GameState, dir zen_doctor.Direction) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		state.MovePlayer(dir)
		v.Clear()
		fmt.Fprintf(v, "%s", state.String())
		return nil
	}
}

func gameOver(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	gameOverText := zen_doctor.GameOver()
	x1, y1, x2, y2 := zen_doctor.CalculateViewPosition(len(gameOverText)+1, 2, maxX, maxY)
	if v, err := g.SetView("game over", x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetCurrentView("game over")
		g.SelFgColor = gocui.ColorRed
		v.Title = "GAME OVER"
		fmt.Fprintf(v, "%s", gameOverText)
	}
	return nil
}

func gameLoop(g *gocui.Gui, state *zen_doctor.GameState) {
	level := state.GetLevel()

	viewUpdate := time.NewTicker(time.Duration(1000/level.FPS) * time.Millisecond)
	defer viewUpdate.Stop()

	fixedUpdate := time.NewTicker((1000 / 30) * time.Millisecond)
	defer fixedUpdate.Stop()

	for {
		select {
		case <-done:
			return

		// Fixed update
		case <-fixedUpdate.C:
			state.TickPlayer()
			g.Update(func(g *gocui.Gui) error {
				// threat view
				if v, err := g.View(threatView); err == nil {
					v.Clear()
					fmt.Fprintf(v, "%s", state.ThreatMeter())
				}
				if state.IsGameOver() {
					done <- true
					return gameOver(g)
				}
				return nil
			})

		// Game time update
		case <-viewUpdate.C:
			// game tick
			state.TickBitStream()

			// update the views
			g.Update(func(g *gocui.Gui) error {
				// main game view
				if v, err := g.View(level.Name()); err == nil {
					v.Clear()
					fmt.Fprintf(v, "%s", state.String())
				}
				return nil
			})
		}
	}
}
