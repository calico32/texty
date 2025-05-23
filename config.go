package texty

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	layershell "github.com/diamondburned/gotk-layer-shell"
	"github.com/gotk3/gotk3/gtk"
)

type Config struct {
	Styles   string    `json:"styles"`
	Windows  []*Window `json:"window"`
	Defaults Window    `json:"defaults"`
}

type Window struct {
	Id       string            `json:"id"`
	Command  []string          `json:"command"`
	Text     *string           `json:"text"`
	File     *string           `json:"file"`
	Interval *TimeSpec         `json:"interval"`
	Position *Position         `json:"position"`
	Layer    *layershell.Layer `json:"layer"`
	Style    *Style            `json:"style"`
	Align    *gtk.Align
}

type Position struct {
	Top    *int `json:"top"`
	Bottom *int `json:"bottom"`
	Left   *int `json:"left"`
	Right  *int `json:"right"`
	Center bool `json:",arg"`
}

type Style struct {
	String string            `json:"string"`
	Map    map[string]string `json:"map"`
}

type TimeSpec time.Duration

func (c *Config) GenerateCSS(verbose bool) (string, error) {
	var styles strings.Builder

	if c.Styles != "" {
		if verbose {
			log.Printf("loading CSS from %s", c.Styles)
		}
		css, err := os.ReadFile(c.Styles)
		if err != nil {
			return "", fmt.Errorf("failed to read styles file: %w", err)
		}
		styles.Write(css)
		styles.WriteString("\n")
	}

	for _, w := range c.Windows {
		if w.Style != nil {
			css := w.GenerateCSS()
			if css != "" {
				styles.WriteString(css)
				styles.WriteString("\n")
			}
		}
	}

	return styles.String(), nil
}

func (w *Window) GenerateCSS() string {
	if w.Style == nil {
		return ""
	}
	if w.Style.String != "" {
		return fmt.Sprintf("#%s {\n%s\n}", w.Id, w.Style.String)
	}
	css := fmt.Sprintf("#%s {\n", w.Id)
	for k, v := range w.Style.Map {
		css += "  " + k + ": " + v + ";\n"
	}
	css += "}\n"
	return css
}

func (c *Config) SerializeJSON() string {
	prefix := "|   "
	j, err := json.MarshalIndent(c, prefix, "  ")
	if err != nil {
		return fmt.Sprintf("Error serializing config to JSON: %v", err)
	}
	return prefix + string(j)
}
