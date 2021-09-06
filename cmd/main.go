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
	done          = make(chan bool)
	selectedLevel = zen_doctor.Tutorial
)

const (
	menuView       = "menu"
	threatView     = "threat"
	threatViewSize = 50
)

func main() {
	rand.Seed(time.Now().Unix())
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SetManagerFunc(layout)
	// set the menu as the initial view
	g.Update(func(gui *gocui.Gui) error {
		_, err := g.SetCurrentView(menuView)
		return err
	})

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	width, height := 50, 20
	if v, err := g.SetView(menuView, maxX/2-width/2, maxY/2-height/2, maxX/2+width/2, maxY/2+height/2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Choose level"
		renderMenu(v)
	}

	if v, err := g.SetView("help", maxX-25, 0, maxX-1, 9); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Help"
		v.Wrap = true
		fmt.Fprintln(v, "Keys:")
		fmt.Fprintln(v, "← ↑ → ↓: Move")
		fmt.Fprintln(v, "^C:      Exit")
		fmt.Fprintln(v, "Objective:")
		fmt.Fprintln(v, "Find required data,\nthen move to exit.")
	}

	if v, err := g.SetView(threatView, 0, 0, threatViewSize, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Threat"
	}
	return nil
}

func renderMenu(v *gocui.View) {
	v.Clear()
	for i, level := range zen_doctor.Levels() {
		if selectedLevel.Equals(i) {
			fmt.Fprintf(v, "> %s\n", zen_doctor.WithColor(zen_doctor.Green, level.String()))
		} else {
			fmt.Fprintf(v, "  %s\n", level.String())
		}
	}
	// 256-colors escape codes
	//for i := 0; i < 256; i++ {
	//	str := fmt.Sprintf("\x1b[48;5;%dm\x1b[30m%3d\x1b[0m ", i, i)
	//	str += fmt.Sprintf("\x1b[38;5;%dm%3d\x1b[0m ", i, i)
	//
	//	if (i+1)%10 == 0 {
	//		str += "\n"
	//	}
	//
	//	fmt.Fprint(v, str)
	//}
	//
	//fmt.Fprint(v, "\n\n")
}

func menuUp(_ *gocui.Gui, v *gocui.View) error {
	if selectedLevel.Dec().IsValid() {
		selectedLevel = selectedLevel.Dec()
	}
	renderMenu(v)
	return nil
}

func menuDown(_ *gocui.Gui, v *gocui.View) error {
	if selectedLevel.Inc().IsValid() {
		selectedLevel = selectedLevel.Inc()
	}
	renderMenu(v)
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(menuView, gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(menuView, gocui.KeyArrowUp, gocui.ModNone, menuUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(menuView, gocui.KeyArrowDown, gocui.ModNone, menuDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(menuView, gocui.KeyEnter, gocui.ModNone, loadGame); err != nil {
		log.Panicln(err)
	}
	return nil
}

func loadGame(g *gocui.Gui, v *gocui.View) error {
	// load this level and close the menu
	level := zen_doctor.GetLevel(selectedLevel)
	if !level.IsValid() {
		fmt.Fprintf(v, "that level is not implemented yet :( \n")
		return nil
	}

	maxX, maxY := g.Size()
	x1, y1, x2, y2 := zen_doctor.CalculateViewPosition(level.Width, level.Height, maxX, maxY)
	levelView, err := g.SetView(level.Name(), x1, y1, x2, y2)
	if err != nil && err != gocui.ErrUnknownView {
		return errors.Wrapf(err, "setting view for level %s", level.Name())
	}
	levelView.Title = level.Name()

	state := zen_doctor.NewGameState(selectedLevel)
	go gameLoop(g, &state)
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}

// callback hell :notlikethis:
func endGame(state *zen_doctor.GameState) func(g *gocui.Gui, _ *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		g.Update(func(g *gocui.Gui) error {
			// go back to menu view
			g.SetCurrentView(menuView)

			// send to the done channel so the goroutine stops
			done <- true

			// delete the data for the level
			level := state.GetLevel()
			g.DeleteView(level.Name())
			g.DeleteKeybindings(level.Name())
			state.Reset()

			// reset threat meter view
			if v, err := g.View(threatView); err == nil {
				v.Clear()
			}
			return nil
		})
		return nil
	}
}

func movePlayer(state *zen_doctor.GameState, dir zen_doctor.Direction) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		state.MovePlayer(dir)
		v.Clear()
		fmt.Fprintf(v, "%s", state.String())
		return nil
	}
}

func gameLoop(g *gocui.Gui, state *zen_doctor.GameState) {
	level := state.GetLevel()

	ticker := time.NewTicker(time.Duration(1000/level.FPS) * time.Millisecond)
	defer ticker.Stop()

	g.Update(func(g *gocui.Gui) error {
		// in-game keybinds
		if err := g.SetKeybinding(level.Name(), gocui.KeyCtrlC, gocui.ModNone, endGame(state)); err != nil {
			return err
		}
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
		v, err := g.SetCurrentView(level.Name())
		if err != nil {
			return err
		}
		v.Clear()
		fmt.Fprintf(v, "%s", state.String())
		return nil
	})

	for {
		select {
		case <-done:
			return

		// player threat is separate from game state
		case <-time.After((1000 / 30) * time.Millisecond):
			state.TickPlayer()
			g.Update(func(g *gocui.Gui) error {
				// threat view
				if v, err := g.View(threatView); err == nil {
					v.Clear()
					fmt.Fprintf(v, "%s", state.ThreatMeter(threatViewSize))
				}
				if state.IsGameOver() {
					ticker.Stop()
					if v, err := g.View(level.Name()); err == nil {
						v.Clear() // TODO: i dunno how to leave the screen and also show this game over message. Maybe a new view?
						fmt.Fprintf(v, "%s", zen_doctor.GameOver())
					}
				}
				return nil
			})

		case <-ticker.C:
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
