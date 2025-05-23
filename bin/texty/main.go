package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"texty"
	"time"

	"log"

	layershell "github.com/diamondburned/gotk-layer-shell"
	"github.com/fsnotify/fsnotify"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"rsc.io/getopt"
)

var configPathFlag = flag.String("config", "", "Path to the config file")
var verboseFlag = flag.Bool("verbose", false, "Enable verbose logging")

var restart = false

func init() {
	getopt.Alias("c", "config")
	getopt.Alias("v", "verbose")
}

type app struct {
	config     texty.Config
	windows    []*window
	stylesheet string
}

type window struct {
	closed   bool
	config   *texty.Window
	window   *gtk.Window
	box      *gtk.Box
	maxWidth int
}

func main() {
	flag.CommandLine.Init("", flag.ExitOnError)
	getopt.Parse()

	verbose := *verboseFlag

	if verbose {
		log.Print("texty initializing")
	}

	gtk.Init(nil)

	var app app
	var configPath string
	var err error

	app.config, configPath, err = texty.LoadConfig(configPathFlag, *verboseFlag)
	if err != nil {
		log.Printf("warning: failed to load config: %v", err)
	}

	go watchConfig(configPath, verbose)

	cssProvider, err := gtk.CssProviderNew()
	if err != nil {
		log.Printf("warning: failed to create CSS provider: %v", err)
		return
	}

	if app.config.Styles != "" {
		if verbose {
			log.Printf("loading CSS from %s", app.config.Styles)
		}
		s, err := os.ReadFile(app.config.Styles)
		if err != nil {
			log.Printf("warning: failed to read CSS file: %v", err)
		}
		app.stylesheet = string(s)
	}

	for _, w := range app.config.Windows {
		css := w.SerializeCSS()
		if css != "" {
			app.stylesheet += css
		}
	}

	if err := cssProvider.LoadFromData(app.stylesheet); err != nil {
		e := fmt.Errorf("failed to load CSS: %w", err)
		config := texty.MakeErrorConfig(e)
		app.config = config
		app.stylesheet = config.Windows[0].SerializeCSS()
		var err error
		cssProvider, err = gtk.CssProviderNew()
		if err != nil {
			log.Fatalf("fatal: failed to create CSS provider for error config: %v", err)
		}
		if err := cssProvider.LoadFromData(app.stylesheet); err != nil {
			log.Fatalf("fatal: failed to load error config CSS: %v", err)
		}
	}

	app.windows = make([]*window, len(app.config.Windows))

	for i, windowConfig := range app.config.Windows {
		w := &window{
			config: windowConfig,
		}

		if verbose {
			log.Printf("creating window %s", windowConfig.Id)
		}

		w.window, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
		if err != nil {
			log.Printf("warning: failed to create window: %v", err)
			continue
		}

		layershell.InitForWindow(w.window)
		layershell.SetLayer(w.window, *w.config.Layer)
		layershell.SetExclusiveZone(w.window, 0)

		if w.config.Position != nil {
			if w.config.Position.Top != nil {
				layershell.SetAnchor(w.window, layershell.EdgeTop, true)
				layershell.SetMargin(w.window, layershell.EdgeTop, *w.config.Position.Top)
			}
			if w.config.Position.Bottom != nil {
				layershell.SetAnchor(w.window, layershell.EdgeBottom, true)
				layershell.SetMargin(w.window, layershell.EdgeBottom, *w.config.Position.Bottom)
			}
			if w.config.Position.Left != nil {
				layershell.SetAnchor(w.window, layershell.EdgeLeft, true)
				layershell.SetMargin(w.window, layershell.EdgeLeft, *w.config.Position.Left)
			}
			if w.config.Position.Right != nil {
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

		w.window.SetProperty("name", windowConfig.Id)

		w.box, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		if err != nil {
			log.Printf("warning: failed to create box: %v", err)
			continue
		}
		w.box.SetMarginTop(4)
		w.box.SetMarginBottom(4)

		w.window.Add(w.box)

		app.windows[i] = w
	}

	for _, w := range app.windows {
		styleContext, err := w.window.GetStyleContext()
		if err != nil {
			log.Printf("warning: failed to get style context: %v", err)
			continue
		}
		styleContext.AddProvider(cssProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)

		w.draw()
	}

	for _, w := range app.windows {
		var h1 glib.SignalHandle
		h1 = w.window.Connect("destroy", func() {
			w.closed = true
			allClosed := true
			for _, w := range app.windows {
				if !w.closed {
					allClosed = false
					break
				}
			}
			if allClosed {
				gtk.MainQuit()
			}
			w.window.HandlerDisconnect(h1)
		})

		// middle mouse button to close
		var h2 glib.SignalHandle
		h2 = w.window.Connect("button-press-event", func(_ *gtk.Window, e *gdk.Event) {
			ev := gdk.EventButtonNewFromEvent(e)
			if ev.Button() == gdk.BUTTON_MIDDLE {
				w.window.Close()
				w.closed = true
				w.window.HandlerDisconnect(h2)
			}
		})

		w.window.ShowAll()

		if w.config.Interval != nil {
			ms := time.Duration(*w.config.Interval) / time.Millisecond
			glib.TimeoutAdd(uint(ms), func() bool {
				w.draw()
				return true
			})
		}
	}

	log.Print("texty started")
	gtk.Main()

	if restart {
		exe, err := os.Executable()
		if err != nil {
			log.Fatal("fatal: couldn't locate the current executable")
		}
		// restart the app
		if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
			log.Fatalf("fatal: failed to restart texty: %v", err)
		}
	}
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
	w.box.GetChildren().Foreach(func(item any) {
		item.(gtk.IWidget).ToWidget().Destroy()
	})

	for _, lineText := range lines {
		line, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		if err != nil {
			log.Printf("warning: failed to create line: %v", err)
			continue
		}
		w.box.PackStart(line, true, false, 0)

		label, err := gtk.LabelNew("")
		if err != nil {
			log.Printf("warning: failed to create label: %v", err)
			continue
		}
		label.SetMarkup(lineText)
		label.SetMarginTop(4)
		label.SetMarginBottom(4)
		line.PackStart(label, true, true, 8)
		if w.config.Position != nil && w.config.Position.Center {
			label.SetHAlign(gtk.ALIGN_CENTER)
			label.SetVAlign(gtk.ALIGN_CENTER)
		} else {
			label.SetHAlign(gtk.ALIGN_START)
			label.SetVAlign(gtk.ALIGN_START)
		}
	}

	w.box.ShowAll()

	allocWidth := w.box.GetAllocatedWidth()
	if allocWidth > w.maxWidth {
		w.maxWidth = allocWidth
		w.window.SetSizeRequest(w.maxWidth, -1)
	}
}

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
