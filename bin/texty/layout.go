package main

import (
	"log"

	layershell "github.com/diamondburned/gotk-layer-shell"
)

func (w *window) layout() {
	layershell.InitForWindow(w.window)
	layershell.SetLayer(w.window, *w.config.Layer)
	layershell.SetExclusiveZone(w.window, 0)

	if w.config.Position != nil {
		if w.config.Position.Top != nil {
			layershell.SetAnchor(w.window, layershell.EdgeTop, true)
			layershell.SetMargin(w.window, layershell.EdgeTop, *w.config.Position.Top)
		} else if w.config.Position.Bottom != nil {
			layershell.SetAnchor(w.window, layershell.EdgeBottom, true)
			layershell.SetMargin(w.window, layershell.EdgeBottom, *w.config.Position.Bottom)
		}

		if w.config.Position.Left != nil {
			layershell.SetAnchor(w.window, layershell.EdgeLeft, true)
			layershell.SetMargin(w.window, layershell.EdgeLeft, *w.config.Position.Left)
		} else if w.config.Position.Right != nil {
			layershell.SetAnchor(w.window, layershell.EdgeRight, true)
			layershell.SetMargin(w.window, layershell.EdgeRight, *w.config.Position.Right)
		}

		if w.config.Position.Center {
			affected := false
			if w.config.Position.Top == nil && w.config.Position.Bottom == nil {
				// center vertically
				layershell.SetAnchor(w.window, layershell.EdgeTop, true)
				layershell.SetAnchor(w.window, layershell.EdgeBottom, true)
				affected = true
			}

			if w.config.Position.Left == nil && w.config.Position.Right == nil {
				// center horizontally
				layershell.SetAnchor(w.window, layershell.EdgeLeft, true)
				layershell.SetAnchor(w.window, layershell.EdgeRight, true)
				affected = true
			}

			if !affected {
				log.Printf("warning: center position doesn't do anything with both vertical and horizontal anchors set")
			}
		}
	}
}
