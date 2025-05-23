package main

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/gotk3/gotk3/gtk"
)

func watchConfig(path string, verbose bool) {
	if path == "" {
		return
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("warning: failed to create watcher: %v", err)
		return
	}
	defer w.Close()

	err = w.Add(filepath.Dir(path))
	if err != nil {
		log.Printf("warning: failed to add watcher: %v", err)
		return
	}

	if verbose {
		log.Printf("watching config file: %s", path)
	}

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Name == path && event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("config file changed: %s", event.Name)
				restart = true
				gtk.MainQuit()
				return
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}
