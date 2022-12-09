package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"gos3/config"
	"gos3/pkg/ceph"
	"log"
	"os"
)

func init() {
	err := ceph.InitAwsClient(config.OSSConfig{
		Endpoint:                  "10.198.30.156",
		AccessKeyID:               "spe",
		AccessKeySecret:           "Spe2077@#$%",
		Bucket:                    "spe_data",
		S3ForcePathStyle:          true,
		InsecureSkipVerify:        true,
		DisableEndpointHostPrefix: true,
	})
	if err != nil {
		panic(err)
	}
	logFile, _ := os.Create("/tmp/gos3.log")
	log.SetOutput(logFile)
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("main", 1, 1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		//v.Wrap = true
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		//res, err := ceph.Client.ListBucket()
		//if err != nil {
		//	log.Panicln(err)
		//}
		for _, b := range []string{"aaa", "bbb", "ccc", "ddd"} {
			fmt.Fprintf(v, "%v\n", b)
		}
		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := initKeyBinding(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func initKeyBinding(g *gocui.Gui) (err error) {
	if err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return
	}

	if err = g.SetKeybinding("main", gocui.KeyEnter, gocui.ModNone, showMsg); err != nil {
		return
	}

	if err = g.SetKeybinding("msg", gocui.KeyCtrlQ, gocui.ModNone, delMsg); err != nil {
		return
	}

	if err = g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return
	}
	if err = g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return
	}

	return
}

func showMsg(g *gocui.Gui, view *gocui.View) error {
	log.Println("show msg")
	if _, e := g.SetCurrentView(view.Name()); e != nil {
		return e
	}

	_, cy := view.Cursor()
	l, _ := view.Line(cy)

	maxX, maxY := g.Size()
	if v, err := g.SetView("msg", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, l)
		if _, err := g.SetCurrentView("msg"); err != nil {
			return err
		}
	}

	return nil
}

func delMsg(g *gocui.Gui, v *gocui.View) error {
	log.Println("del msg")
	if err := g.DeleteView("msg"); err != nil {
		return err
	}
	_, err := g.SetCurrentView("main")
	if err != nil {
		return err
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}
