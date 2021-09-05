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

const menuView = "menu"

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
		fmt.Fprintln(v, "Keys")
		fmt.Fprintln(v, "← ↑ → ↓: Move")
		fmt.Fprintln(v, "^C:      Exit")
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
	//// 256-colors escape codes
	//for i := 100; i < 256; i++ {
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

	go gameLoop(g, selectedLevel)
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}

// callback hell :notlikethis:
func endGame(level zen_doctor.LevelSettings) func(g *gocui.Gui, _ *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		g.Update(func(g *gocui.Gui) error {
			// go back to menu view
			g.SetCurrentView(menuView)

			// send to the done channel so the goroutine stops
			done <- true

			// delete the data for the level
			g.DeleteView(level.Name())
			g.DeleteKeybindings(level.Name())
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

func gameLoop(g *gocui.Gui, lvl zen_doctor.Level) {
	ticker := time.NewTicker((1000 / 5) * time.Millisecond)
	defer ticker.Stop()

	state := zen_doctor.NewGameState(lvl)
	level := zen_doctor.GetLevel(state.CurrentLevel)

	g.Update(func(g *gocui.Gui) error {
		// in-game keybinds
		if err := g.SetKeybinding(level.Name(), gocui.KeyCtrlC, gocui.ModNone, endGame(level)); err != nil {
			return err
		}
		if err := g.SetKeybinding(level.Name(), gocui.KeyArrowUp, gocui.ModNone, movePlayer(&state, zen_doctor.MoveUp)); err != nil {
			return err
		}
		if err := g.SetKeybinding(level.Name(), gocui.KeyArrowDown, gocui.ModNone, movePlayer(&state, zen_doctor.MoveDown)); err != nil {
			return err
		}
		if err := g.SetKeybinding(level.Name(), gocui.KeyArrowLeft, gocui.ModNone, movePlayer(&state, zen_doctor.MoveLeft)); err != nil {
			return err
		}
		if err := g.SetKeybinding(level.Name(), gocui.KeyArrowRight, gocui.ModNone, movePlayer(&state, zen_doctor.MoveRight)); err != nil {
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
		case <-ticker.C:
			// game tick
			g.Update(func(g *gocui.Gui) error {
				v, err := g.View(level.Name())
				if err != nil {
					return err
				}
				v.Clear()
				state.Tick()
				fmt.Fprintf(v, "%s", state.String())

				return nil
			})
		}
	}
}
