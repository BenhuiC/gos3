package widget

import "github.com/jroimartin/gocui"

type Win interface {
	Name() string
	Show(g *gocui.Gui) (err error)
	BindKey(g *gocui.Gui) (err error)
	Quit(g *gocui.Gui) (err error)
	Response() interface{}
}

type eventHandler func(gui *gocui.Gui, view *gocui.View) error
