package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	zen_doctor "github.com/krixi/zen-doctor/internal"
	"github.com/pkg/errors"
)

const (
	threatView      = "threat"
	progressBarView = "progress"
	itemsView       = "items"
)

var (
	done = make(chan bool)
	// we need this to be global so we can replace it when the level is over.
	state     zen_doctor.GameState
	collected = make([]zen_doctor.Loot, 0)
	lastTick  = time.Now()
	elapsed   = 0 * time.Millisecond
	mode      = zen_doctor.CompatibilityAny
	cheatMode = false
)

func main() {
	rand.Seed(time.Now().Unix())
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	mode = parseArgs()
	if mode == zen_doctor.CompatibilityAscii {
		g.ASCII = true
	}
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen

	state = zen_doctor.NewGameState(zen_doctor.Tutorial, mode)
	if err := initGame(g, &state); err != nil {
		log.Panicln(err)
	}

	// start the terminal display loop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func parseArgs() zen_doctor.CompatibilityMode {
	mode := zen_doctor.CompatibilityAny
	for _, arg := range os.Args {
		switch strings.ToLower(arg) {
		case "--ascii":
			mode = zen_doctor.CompatibilityAscii
		case "--latin":
			mode = zen_doctor.CompatibilityLatin
		}
	}
	return mode
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}

func initGame(g *gocui.Gui, state *zen_doctor.GameState) error {
	// reset the layout manager - this creates the view
	g.SetManagerFunc(layout(state))

	// global ket to quit
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	// global keybinds to switch levels
	g.SetKeybinding("", gocui.KeyCtrlY, gocui.ModNone, skipToLevel(zen_doctor.Level1))
	g.SetKeybinding("", gocui.KeyCtrlU, gocui.ModNone, skipToLevel(zen_doctor.Level2))
	g.SetKeybinding("", gocui.KeyCtrlI, gocui.ModNone, skipToLevel(zen_doctor.Level3))
	g.SetKeybinding("", gocui.KeyCtrlO, gocui.ModNone, skipToLevel(zen_doctor.Level4))
	g.SetKeybinding("", gocui.KeyCtrlP, gocui.ModNone, skipToLevel(zen_doctor.Level5))

	// game-specific keybinds
	if err := gameKeybinds(g, state); err != nil {
		return err
	}
	// start the game loop
	go gameLoop(g, state)

	return nil
}

func layout(state *zen_doctor.GameState) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		level := state.Level()
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
		if v, err := g.SetView(itemsView, x1-20, y1-3, x1-1, y2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Items"
			v.Wrap = true
			renderInventory(v, state)
		}
		if v, err := g.SetView(progressBarView, x1, y2+1, x2, y2+3); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = ""
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
func renderInventory(v *gocui.View, state *zen_doctor.GameState) {
	v.Clear()
	fmt.Fprintln(v, "Want:")
	fmt.Fprintf(v, state.DataWanted())
	fmt.Fprintf(v, strings.Repeat("???", 18))
	fmt.Fprintln(v, "Have:")
	fmt.Fprintf(v, state.DataCollected())
	fmt.Fprintf(v, strings.Repeat("???", 18))
	fmt.Fprintln(v, "Collected:")
	b := strings.Builder{}
	for _, have := range state.Inventory() {
		b.WriteString(zen_doctor.WithColor(have.SymbolForMode(mode)))
	}
	fmt.Fprintln(v, b.String())
	fmt.Fprintf(v, strings.Repeat("???", 18))
	fmt.Fprintf(v, zen_doctor.ElapsedTime(elapsed))
}

func gameKeybinds(g *gocui.Gui, state *zen_doctor.GameState) error {
	level := state.Level()
	// in-game keybinds:
	// up
	if err := g.SetKeybinding(level.Name(), gocui.KeyArrowUp, gocui.ModNone, movePlayer(state, zen_doctor.MoveUp)); err != nil {
		return err
	}
	if err := g.SetKeybinding(level.Name(), 'w', gocui.ModNone, movePlayer(state, zen_doctor.MoveUp)); err != nil {
		return err
	}

	// down
	if err := g.SetKeybinding(level.Name(), gocui.KeyArrowDown, gocui.ModNone, movePlayer(state, zen_doctor.MoveDown)); err != nil {
		return err
	}
	if err := g.SetKeybinding(level.Name(), 's', gocui.ModNone, movePlayer(state, zen_doctor.MoveDown)); err != nil {
		return err
	}

	// left
	if err := g.SetKeybinding(level.Name(), gocui.KeyArrowLeft, gocui.ModNone, movePlayer(state, zen_doctor.MoveLeft)); err != nil {
		return err
	}
	if err := g.SetKeybinding(level.Name(), 'a', gocui.ModNone, movePlayer(state, zen_doctor.MoveLeft)); err != nil {
		return err
	}

	// right
	if err := g.SetKeybinding(level.Name(), gocui.KeyArrowRight, gocui.ModNone, movePlayer(state, zen_doctor.MoveRight)); err != nil {
		return err
	}
	if err := g.SetKeybinding(level.Name(), 'd', gocui.ModNone, movePlayer(state, zen_doctor.MoveRight)); err != nil {
		return err
	}

	// pause
	if err := g.SetKeybinding(level.Name(), gocui.KeySpace, gocui.ModNone, pause); err != nil {
		return err
	}
	return nil
}

func movePlayer(state *zen_doctor.GameState, dir zen_doctor.Direction) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		state.MovePlayer(dir)
		v.Clear()
		fmt.Fprintf(v, "%s", state.String())
		return nil
	}
}

// kills the game's goroutine and shows a pause window
func pause(g *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := g.Size()
	x1, y1, x2, y2 := zen_doctor.CalculateViewPosition(80, 2, maxX, maxY)
	if v, err := g.SetView("pause", x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetCurrentView("pause")
		done <- true
		v.Title = "Paused"
		fmt.Fprintln(v, "Press space to resume")

		if err := g.SetKeybinding("pause", gocui.KeySpace, gocui.ModNone, resume); err != nil {
			return err
		}
	}
	return nil
}

func resume(g *gocui.Gui, _ *gocui.View) error {
	g.SetCurrentView(state.Level().Name())
	g.DeleteView("pause")
	g.DeleteKeybindings("pause")
	lastTick = time.Now()
	go gameLoop(g, &state)
	return nil
}

func gameOver(g *gocui.Gui, didWin bool) error {
	// copy over inventory to our final collection
	collected = append(collected, state.Inventory()...)

	maxX, maxY := g.Size()
	gameOverText := zen_doctor.GameOver(didWin, elapsed, mode, collected...)
	x1, y1, x2, y2 := zen_doctor.CalculateViewPosition(80, 8, maxX, maxY)
	if v, err := g.SetView("game over", x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetCurrentView("game over")
		if cheatMode {
			g.SelFgColor = gocui.ColorYellow
			v.Title = "YOU CHEATED"
		} else if didWin {
			g.SelFgColor = gocui.ColorGreen
			v.Title = "YOU WIN"
		} else {
			g.SelFgColor = gocui.ColorRed
			v.Title = "GAME OVER"
		}
		fmt.Fprintf(v, "%s", gameOverText)
		if err := g.SetKeybinding("game over", gocui.KeySpace, gocui.ModNone, restart); err != nil {
			return err
		}
	}
	return nil
}

func restart(g *gocui.Gui, _ *gocui.View) error {
	g.DeleteView("game over")
	g.DeleteKeybindings("game over")
	g.SelFgColor = gocui.ColorGreen
	lastTick = time.Now()
	elapsed = 0 * time.Millisecond
	cheatMode = false
	state = zen_doctor.NewGameState(zen_doctor.Tutorial, mode)
	if err := initGame(g, &state); err != nil {
		return err
	}
	return nil
}

func nextLevel(g *gocui.Gui) error {
	// keep going until they run out of levels - if they make it all the way, winner winner chicken dinner!
	current := state.Level()
	next := current.Level.Inc()
	if !next.IsValid() {
		return gameOver(g, true)
	}
	// clean up old view
	g.DeleteKeybindings(current.Name())
	g.DeleteView(current.Name())

	// copy over inventory to our final collection
	collected = append(collected, state.Inventory()...)

	// create new state and initialize
	loc := state.PlayerLocation()
	state = zen_doctor.NewGameStateWithPlayerAt(loc, next, mode)
	return initGame(g, &state)
}

func skipToLevel(level zen_doctor.Level) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		// clean up the current level and set cheat mode to true to tattle on them at the end
		done <- true
		cheatMode = true
		current := state.Level()
		g.DeleteKeybindings(current.Name())
		g.DeleteView(current.Name())

		// initialize the requested level
		state = zen_doctor.NewGameState(level, mode)
		return initGame(g, &state)
	}
}

func gameLoop(g *gocui.Gui, state *zen_doctor.GameState) {
	level := state.Level()

	viewUpdate := time.NewTicker(time.Duration(1000/level.FPS) * time.Millisecond)
	defer viewUpdate.Stop()

	fixedUpdate := time.NewTicker((1000 / 30) * time.Millisecond)
	defer fixedUpdate.Stop()

	animationUpdate := time.NewTicker((1000 / 11) * time.Millisecond)
	defer animationUpdate.Stop()

	automoveUpdate := time.NewTicker((1000 / 7) * time.Millisecond)
	defer automoveUpdate.Stop()

	for {
		select {
		case <-done:
			return

		// Fixed update
		case <-fixedUpdate.C:
			now := time.Now()
			elapsed = elapsed + now.Sub(lastTick)
			lastTick = now

			state.TickWorld()
			state.TickPlayer()
			g.Update(func(g *gocui.Gui) error {
				// threat view
				if v, err := g.View(threatView); err == nil {
					v.Clear()
					fmt.Fprintf(v, "%s", state.ThreatMeter())
				}
				// loot view
				if v, err := g.View(progressBarView); err == nil {
					v.Clear()
					v.Title = state.ProgressBarType()
					fmt.Fprintf(v, "%s", state.ProgressBar())
				}
				// inventory view
				if v, err := g.View(itemsView); err == nil {
					renderInventory(v, state)
				}
				// main game view
				if v, err := g.View(level.Name()); err == nil {
					v.Clear()
					fmt.Fprintf(v, "%s", state.String())
				}
				if state.IsComplete() {
					done <- true
					return nextLevel(g)
				}
				if state.IsGameOver() {
					done <- true
					return gameOver(g, false)
				}
				return nil
			})

		// Game time update
		case <-viewUpdate.C:
			state.TickBitStream()

		// animations run at a different speed
		case <-animationUpdate.C:
			state.TickAnimations()

			// automatic movement
		case <-automoveUpdate.C:
			state.TickMovement()
		}
	}
}
