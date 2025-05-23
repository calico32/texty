package main

import (
	"log"
	"texty"

	"github.com/gotk3/gotk3/gtk"
)

type window struct {
	closed     bool
	config     *texty.Window
	window     *gtk.Window
	container  *gtk.Box
	contentBox *gtk.Box
	maxWidth   int
}

func newWindow(config *texty.Window, verbose bool) (*window, error) {
	w := &window{config: config}
	var err error

	if verbose {
		log.Printf("creating window %s", config.Id)
	}

	w.window, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Printf("error: failed to create window: %v", err)
		return nil, err
	}

	w.window.SetDecorated(false)
	err = w.window.SetProperty("name", config.Id)
	if err != nil {
		log.Printf("error: failed to set window name: %v", err)
		return nil, err
	}

	w.container, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		log.Printf("error: failed to create box: %v", err)
		return nil, err
	}
	w.container.SetHAlign(gtk.ALIGN_CENTER)
	w.container.SetVAlign(gtk.ALIGN_CENTER)
	w.window.Add(w.container)

	w.contentBox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		log.Printf("error: failed to create box: %v", err)
		return nil, err
	}
	w.contentBox.SetMarginTop(4)
	w.contentBox.SetMarginBottom(4)
	w.contentBox.SetHAlign(gtk.ALIGN_CENTER)
	w.contentBox.SetVAlign(gtk.ALIGN_CENTER)
	w.container.Add(w.contentBox)

	w.layout()

	return w, nil
}
