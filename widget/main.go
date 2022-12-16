package widget

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"gos3/pkg/util"
)

type Main struct {
	name string
}

func (m *Main) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(m.Name(), 1, 1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		for _, b := range []string{"aaa", "bbb", "ccc", "ddd"} {
			fmt.Fprintf(v, "%v\n", b)
		}
		if _, err := g.SetCurrentView(m.Name()); err != nil {
			return err
		}
	}

	return nil
}

func (m *Main) Name() string {
	if m.name == "" {
		m.name = util.UUID32()
	}
	return fmt.Sprintf("Main_%s", m.name)
}

func (m *Main) BindKey(g *gocui.Gui) (err error) {
	return nil
}

func (m *Main) Quit(g *gocui.Gui) (err error) {
	return g.DeleteView(m.Name())
}

func (m *Main) Response() interface{} {
	return nil
}

func (m *Main) Show(g *gocui.Gui) error {
	return nil
}
