package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gotk3/gotk3/glib"
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
	text, err := w.getText()
	if err != nil {
		log.Printf("warning: failed to get text: %v", err)
		return
	}
	glib.IdleAdd(func() {
		w.updateText(text)
	})
}

func (w *window) updateText(text string) {
	if w.closed {
		return
	}

	lines := strings.Split(text, "\n")
	if len(lines) != 0 {
		// remove empty lines from the beginning and end
		for lines[0] == "" {
			lines = lines[1:]
		}
		for lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	}

	spacing := 8
	if w.config.Spacing != nil {
		spacing = *w.config.Spacing
	}

	glib.TimeoutAdd(0, func() bool {
		w.contentBox.GetChildren().Foreach(func(item any) {
			item.(gtk.IWidget).ToWidget().Destroy()
		})

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

		allocWidth := w.contentBox.GetAllocatedWidth()
		if allocWidth > w.maxWidth {
			w.maxWidth = allocWidth
			w.window.SetSizeRequest(w.maxWidth, -1)
		}

		w.window.ShowAll()

		return false
	})
}

func (w *window) jsonLoop() {

	c := exec.Command(w.config.Command[0], w.config.Command[1:]...)
	out, err := c.StdoutPipe()
	if err != nil {
		log.Printf("warning: failed to get stdout pipe: %v", err)
		return
	}
	err = c.Start()
	if err != nil {
		log.Printf("warning: failed to start command: %v", err)
		return
	}

	r := bufio.NewReader(out)
	defer c.Wait()

	c.Start()
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			log.Printf("warning: failed to read line: %v", err)
			break
		}
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var jsonContent struct {
			Text string `json:"text"`
		}

		if err := json.Unmarshal(line, &jsonContent); err != nil {
			log.Printf("warning: failed to unmarshal JSON: %v", err)
			continue
		}

		if w.closed {
			return
		}

		glib.IdleAdd(func() {
			w.updateText(jsonContent.Text)
		})
	}

}
