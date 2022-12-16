package widget

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"gos3/pkg/util"
)

type Conform struct {
	name    string
	body    string
	yes, no *Button
	res     chan bool
}

func NewConformWin(name, body string) *Conform {
	ch := make(chan bool, 1)
	return &Conform{
		name: name,
		body: body,
		res:  ch,
		yes: NewButton(name+"_yes", 0, 0, "yes", func(g *gocui.Gui, v *gocui.View) error {
			ch <- true
			return nil
		}),
		no: NewButton(name+"_no", 0, 0, "no", func(g *gocui.Gui, v *gocui.View) error {
			ch <- false
			return nil
		}),
	}
}

func (c *Conform) Name() string {
	if c.name == "" {
		c.name = util.UUID32()
	}
	return fmt.Sprintf("ConformWin_%s", c.name)
}

func (c *Conform) BindKey(g *gocui.Gui) (err error) {
	quit := func(gui *gocui.Gui, view *gocui.View) error {
		return c.Quit(gui)
	}
	tab := func(gui *gocui.Gui, view *gocui.View) error {
		nextView := c.yes.Name()
		if view.Name() == nextView {
			nextView = c.no.Name()
		}
		_, err := gui.SetCurrentView(nextView)
		return err
	}
	left := func(gui *gocui.Gui, view *gocui.View) error {
		_, err := gui.SetCurrentView(c.no.Name())
		return err
	}
	right := func(gui *gocui.Gui, view *gocui.View) error {
		_, err := gui.SetCurrentView(c.yes.Name())
		return err
	}
	if err = g.SetKeybinding(c.yes.Name(), gocui.KeyEsc, gocui.ModNone, quit); err != nil {
		return
	}
	if err = g.SetKeybinding(c.no.Name(), gocui.KeyEsc, gocui.ModNone, quit); err != nil {
		return
	}
	if err = g.SetKeybinding(c.Name(), gocui.KeyEsc, gocui.ModNone, quit); err != nil {
		return
	}

	if err = g.SetKeybinding(c.yes.Name(), gocui.KeyTab, gocui.ModNone, tab); err != nil {
		return
	}
	if err = g.SetKeybinding(c.yes.Name(), gocui.KeyArrowRight, gocui.ModNone, right); err != nil {
		return
	}

	if err = g.SetKeybinding(c.no.Name(), gocui.KeyTab, gocui.ModNone, tab); err != nil {
		return
	}
	if err = g.SetKeybinding(c.no.Name(), gocui.KeyArrowRight, gocui.ModNone, left); err != nil {
		return
	}

	return nil
}

func (c *Conform) Quit(g *gocui.Gui) (err error) {
	if err = g.DeleteView(c.yes.Name()); err != nil {
		return
	}
	if err = g.DeleteView(c.no.Name()); err != nil {
		return
	}
	err = g.DeleteView(c.Name())
	return nil
}

func (c *Conform) Response() interface{} {
	r := <-c.res
	return r
}

func (c *Conform) Show(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	width := maxX / 4
	height := maxY / 4
	x0 := maxX/2 - width/2
	y0 := maxY/2 - height/2 - 2
	x1 := maxX/2 + width/2
	y1 := maxY/2 + height/2
	if v, err := g.SetView(c.Name(), x0, y0, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Wrap = true

		//yesView, err := g.SetView(c.yes.Name(), x0+1, y1-3, x0+4, y1-1)
		//if err != nil && err != gocui.ErrUnknownView {
		//	return err
		//}
		//
		//noView, err := g.SetView(c.no.Name(), x1-4, y1-3, x1-1, y1-1)
		//if err != nil && err != gocui.ErrUnknownView {
		//	return err
		//}
		//_, _ = yesView, noView

		c.yes.SetLocation(x0+1, y1-3, 3)
		c.no.SetLocation(x1-4, y1-3, 3)
		if err = c.no.Show(g); err != nil {
			return err
		}
		if err = c.yes.Show(g); err != nil {
			return err
		}

		if err = c.BindKey(g); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(v, "\n\t%v\n", c.body)
		//if _, err := g.SetCurrentView(c.yes.Name()); err != nil {
		//	return err
		//}
	}
	return nil
}
