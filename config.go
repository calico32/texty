package texty

import (
	"encoding/json"
	"fmt"
	"time"

	layershell "github.com/diamondburned/gotk-layer-shell"
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

func (w *Window) SerializeCSS() string {
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
