package main

import (
	"fmt"
	zen_doctor "github.com/krixi/zen-doctor/internal"
	"log"
	"math/rand"
	"time"

	"github.com/jroimartin/gocui"
)

var done = make(chan struct{})

func main() {
	rand.Seed(time.Now().Unix())
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	go gameLoop(g)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func gameLoop(ui *gocui.Gui) {
	ticker := time.NewTicker((1000 / 5) * time.Millisecond)
	defer ticker.Stop()

	state := zen_doctor.NewGameState()
	state.InitView()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			level := zen_doctor.GetLevel(state.CurrentLevel)
			// game tick
			ui.Update(func(g *gocui.Gui) error {
				v, err := g.View("matrix")
				if err != nil {
					return err
				}
				v.Title = level.Name
				v.Clear()

				fmt.Fprintf(v, "%s", state.String())
				state.Shift(1, 1)

				return nil
			})
		}
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("matrix", 1, 1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintf(v, "%s\n", zen_doctor.WithColor(zen_doctor.Yellow, "Loading...\n"))
	}
	//if v, err := g.SetView("colors", 0, 0, maxX, maxY); err != nil {
	//	if err != gocui.ErrUnknownView {
	//		return err
	//	}
	//
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
	//
	//	// 8-colors escape codes
	//	ctr := 0
	//	for i := 0; i <= 7; i++ {
	//		for _, j := range []int{1, 4, 7} {
	//			str := fmt.Sprintf("\x1b[3%d;%dm%d:%d\x1b[0m ", i, j, i, j)
	//			if (ctr+1)%20 == 0 {
	//				str += "\n"
	//			}
	//
	//			fmt.Fprint(v, str)
	//
	//			ctr++
	//		}
	//	}
	//}
	return nil
}

func keybindings(g *gocui.Gui) error {

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}
