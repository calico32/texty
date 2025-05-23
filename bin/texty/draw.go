package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gotk3/gotk3/gtk"
)

func (w *window) getText() (string, error) {
	if w.config.Text != nil {
		return *w.config.Text, nil
	}
	if w.config.File != nil {
		text, err := os.ReadFile(*w.config.File)
		if err != nil {
			return "", err
		}
		return string(text), nil
	}

	cmd, err := exec.Command(w.config.Command[0], w.config.Command[1:]...).Output()
	if err != nil {
		return "", err
	}
	return string(cmd), nil
}

func (w *window) draw() {
	if w.closed {
		return
	}

	text, err := w.getText()
	if err != nil {
		log.Printf("warning: failed to get text: %v", err)
		return
	}

	lines := strings.Split(strings.TrimSpace(text), "\n")
	w.contentBox.GetChildren().Foreach(func(item any) {
		item.(gtk.IWidget).ToWidget().Destroy()
	})

	spacing := 8
	if w.config.Spacing != nil {
		spacing = *w.config.Spacing
	}

	for _, lineText := range lines {
		line, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		if err != nil {
			log.Printf("warning: failed to create line: %v", err)
			continue
		}
		w.contentBox.PackStart(line, true, false, 0)

		label, err := gtk.LabelNew("")
		if err != nil {
			log.Printf("warning: failed to create label: %v", err)
			continue
		}
		label.SetMarkup(lineText)
		label.SetMarginBottom(spacing)
		line.PackStart(label, true, true, 8)
		if w.config.Position != nil && w.config.Position.Center && w.config.Align == nil {
			label.SetHAlign(gtk.ALIGN_CENTER)
			label.SetVAlign(gtk.ALIGN_CENTER)
		} else if w.config.Align == nil {
			label.SetHAlign(gtk.ALIGN_START)
			label.SetVAlign(gtk.ALIGN_START)
		} else {
			label.SetHAlign(*w.config.Align)
			label.SetVAlign(*w.config.Align)
		}
	}

	w.contentBox.ShowAll()

	allocWidth := w.contentBox.GetAllocatedWidth()
	if allocWidth > w.maxWidth {
		w.maxWidth = allocWidth
		w.window.SetSizeRequest(w.maxWidth, -1)
	}
}
