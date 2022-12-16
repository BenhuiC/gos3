package widget

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"gos3/pkg/util"
)

type Button struct {
	name    string
	x, y    int
	w       int
	label   string
	handler func(g *gocui.Gui, v *gocui.View) error
}

func NewButton(name string, x, y int, label string, handler func(g *gocui.Gui, v *gocui.View) error) *Button {
	return &Button{name: name, x: x, y: y, w: len(label) + 1, label: label, handler: handler}
}

func (b *Button) SetLocation(x, y, w int) {
	b.x, b.y, b.w = x, y, w
}

func (b *Button) Name() string {
	if b.name == "" {
		b.name = util.UUID32()
	}
	return fmt.Sprintf("Button_%s", b.name)
}

func (b *Button) BindKey(g *gocui.Gui) (err error) {
	return
}

func (b *Button) Quit(g *gocui.Gui) (err error) {
	return
}

func (b *Button) Response() interface{} {
	return nil
}

func (b *Button) Show(g *gocui.Gui) error {
	v, err := g.SetView(b.Name(), b.x, b.y, b.x+b.w, b.y+2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if _, err := g.SetCurrentView(b.Name()); err != nil {
			return err
		}
		if err := g.SetKeybinding(b.Name(), gocui.KeyEnter, gocui.ModNone, b.handler); err != nil {
			return err
		}
		fmt.Fprint(v, b.label)
	}
	return nil
}
