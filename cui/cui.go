package cui

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/jroimartin/gocui"
)

var g *gocui.Gui
var logger *log.Logger

var connMap map[gopacket.Flow]int

type Key struct {
	viewname string
	key      interface{}
	handler  func(*gocui.Gui, *gocui.View) error
}

var keys []Key = []Key{
	Key{"", gocui.KeyCtrlC, actionGlobalQuit},
	Key{"", gocui.KeyArrowUp, actionGlobalArrowUp},
	Key{"", gocui.KeyArrowDown, actionGlobalArrowDown},
	Key{"", gocui.KeyEnter, actionEnterKey},
	Key{"", gocui.KeyArrowRight, actionEnterKey},
	Key{"", gocui.KeyArrowLeft, actionArrowLeftKey},
}

func InitCui(guiLogger *log.Logger) {
	logger = guiLogger
	connMap = map[gopacket.Flow]int{}
	gui, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		panic(fmt.Sprintf("could not init new gui: %s\n", err.Error()))
	}

	g = gui
	g.Mouse = true

	g.SetManagerFunc(layout)

	for _, key := range keys {
		if err := g.SetKeybinding(key.viewname, key.key, gocui.ModNone, key.handler); err != nil {
			panic(fmt.Sprintf("could not set key bindings: %s\n", err.Error()))
		}
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(fmt.Sprintf("received error in main loop: %s\n", err.Error()))
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("title", -1, -1, maxX, 1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Frame = false
		v.BgColor = gocui.ColorDefault | gocui.AttrReverse
		v.FgColor = gocui.ColorDefault | gocui.AttrReverse

		fmt.Fprintln(v, "⣿ NetPeek")
	}

	if v, err := g.SetView("conns", -1, 1, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Frame = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.SetCursor(0, 2)

		addLineToViewConns(v, "SNo", "SRC", "DST")
		fmt.Fprintln(v, strings.Repeat("─", maxX))

		g.SetCurrentView(v.Name())
	}

	if v, err := g.SetView("status", -1, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Frame = false
		v.BgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite

		updateStatus(g, "C")
	}
	return nil
}

// credits: https://github.com/mephux/komanda-cli/blob/master/komanda/color/color.go
func stringFormatBoth(fg, bg int, str string, args []string) string {
	return fmt.Sprintf("\x1b[48;5;%dm\x1b[38;5;%d;%sm%s\x1b[0m", bg, fg, strings.Join(args, ";"), str)
}

func frameText(text string) string {
	return stringFormatBoth(15, 0, text, []string{"1"})
}
