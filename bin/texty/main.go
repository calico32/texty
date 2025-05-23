package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"texty"
	"time"

	"log"

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

func main() {
	flag.CommandLine.Init("", flag.ExitOnError)
	getopt.Parse()

	verbose := *verboseFlag
	if verbose {
		log.Print("texty initializing")
	}

	gtk.Init(nil)

	config, configPath, err := texty.LoadConfig(configPathFlag, *verboseFlag)
	if err != nil {
		log.Printf("warning: failed to load config: %v", err)
	}

	go watchConfig(configPath, verbose)

	cssProvider, err := gtk.CssProviderNew()
	if err != nil {
		log.Printf("warning: failed to create CSS provider: %v", err)
		return
	}

	stylesheet, err := config.GenerateCSS(verbose)
	if err != nil {
		log.Printf("warning: failed to generate CSS: %v", err)
	}

	if err := cssProvider.LoadFromData(stylesheet); err != nil {
		e := fmt.Errorf("failed to load CSS: %w", err)
		config = texty.MakeErrorConfig(e)
		stylesheet = config.Windows[0].GenerateCSS()
		cssProvider, err = gtk.CssProviderNew()
		if err != nil {
			log.Fatalf("fatal: failed to create CSS provider for error config: %v", err)
		}
		if err := cssProvider.LoadFromData(stylesheet); err != nil {
			log.Fatalf("fatal: failed to load error config CSS: %v", err)
		}
	}

	windows := make([]*window, 0, len(config.Windows))

	for _, windowConfig := range config.Windows {
		w, err := newWindow(windowConfig, verbose)
		if err != nil {
			log.Print(err)
			continue
		}

		styleContext, err := w.window.GetStyleContext()
		if err != nil {
			log.Printf("warning: failed to get style context: %v", err)
			continue
		}
		styleContext.AddProvider(cssProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)

		if w.config.CommandFormat == texty.CommandFormatJson {
			go w.jsonLoop()
		} else {
			go w.draw()
		}

		windows = append(windows, w)
	}

	for _, w := range windows {
		var h1 glib.SignalHandle
		h1 = w.window.Connect("destroy", func() {
			w.closed = true
			allClosed := true
			for _, w := range windows {
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

		if w.config.Interval != nil {
			ms := time.Duration(*w.config.Interval) / time.Millisecond
			glib.TimeoutAdd(uint(ms), func() bool {
				go w.draw()
				return !w.closed
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
