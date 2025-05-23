package texty

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/calico32/genpass"
	"github.com/calico32/kdl-go"
	layershell "github.com/diamondburned/gotk-layer-shell"
	"github.com/gotk3/gotk3/gtk"
)

func (c *Config) UnmarshalKDL(doc *kdl.Document) error {
	c.Windows = make([]*Window, 0)
	for _, node := range doc.Nodes {
		switch node.Name {
		case "styles":
			if len(node.Arguments) != 1 {
				return fmt.Errorf("missing path to CSS file for `styles` property")
			}
			if str, ok := node.Arguments[0].(kdl.String); ok {
				c.Styles = fmt.Sprint(str.Value())
			} else {
				return fmt.Errorf("invalid path to CSS file for `styles` property: %v", node.Arguments[0])
			}
		case "window":
			var window Window
			if err := window.UnmarshalKDL(node); err != nil {
				return fmt.Errorf("failed to unmarshal window: %v", err)
			}
			c.Windows = append(c.Windows, &window)
		case "defaults":
			if err := c.Defaults.UnmarshalKDL(node); err != nil {
				return fmt.Errorf("failed to unmarshal defaults: %v", err)
			}
		default:
			return fmt.Errorf("unknown property: %s", node.Name)
		}
	}

	return nil
}

var layers = map[string]layershell.Layer{
	"top":        layershell.LayerTop,
	"bottom":     layershell.LayerBottom,
	"overlay":    layershell.LayerOverlay,
	"background": layershell.LayerBackground,
}

var alignments = map[string]gtk.Align{
	"left":   gtk.ALIGN_START,
	"center": gtk.ALIGN_CENTER,
	"right":  gtk.ALIGN_END,
}

func (w *Window) UnmarshalKDL(node *kdl.Node) error {
	if id, ok := node.Properties["id"]; ok {
		if str, ok := id.(kdl.String); ok {
			w.Id = fmt.Sprint(str.Value())
		} else {
			return fmt.Errorf("invalid id: %v", id)
		}
	} else {
		w.Id = generateRandomId()
	}

	for _, node := range node.Children {
		switch node.Name {
		case "command":
			w.Command = make([]string, len(node.Arguments))
			if len(node.Arguments) == 0 {
				return fmt.Errorf("command requires at least one argument")
			}
			for i, arg := range node.Arguments {
				w.Command[i] = fmt.Sprint(arg.Value())
			}
		case "text":
			var text strings.Builder
			for i, arg := range node.Arguments {
				if i > 0 {
					text.WriteString(" ")
				}
				text.WriteString(fmt.Sprint(arg.Value()))
			}
			txt := text.String()
			w.Text = &txt
		case "file":
			if str, ok := node.Arguments[0].(kdl.String); ok {
				f := fmt.Sprint(str.Value())
				w.File = &f
			} else {
				return fmt.Errorf("invalid file: %v", node.Arguments[0])
			}
			if len(node.Arguments) > 1 {
				return fmt.Errorf("too many arguments for file: %v", node.Arguments)
			}
		case "interval":
			w.Interval = new(TimeSpec)
			if err := w.Interval.UnmarshalKDL(node); err != nil {
				return fmt.Errorf("invalid interval: %v", err)
			}
		case "position":
			w.Position = new(Position)
			if err := w.Position.UnmarshalKDL(node); err != nil {
				return fmt.Errorf("invalid position: %v", err)
			}
		case "layer":
			if str, ok := node.Arguments[0].(kdl.String); ok {
				if layer, ok := layers[fmt.Sprint(str.Value())]; ok {
					w.Layer = &layer
				} else {
					return fmt.Errorf("invalid layer: %s", str.Value())
				}
			} else {
				return fmt.Errorf("invalid layer: %v", node.Arguments[0])
			}
			if len(node.Arguments) > 1 {
				return fmt.Errorf("too many arguments for layer: %v", node.Arguments)
			}
		case "style":
			w.Style = new(Style)
			if err := w.Style.UnmarshalKDL(node); err != nil {
				return fmt.Errorf("invalid style: %v", err)
			}
		case "align":
			if str, ok := node.Arguments[0].(kdl.String); ok {
				if align, ok := alignments[fmt.Sprint(str.Value())]; ok {
					w.Align = &align
				} else {
					return fmt.Errorf("invalid align: %s", str.Value())
				}
			}
		default:
			return fmt.Errorf("unknown property: %s", node.Name)
		}
	}

	return nil
}

func generateRandomId() string {
	id := genpass.Generate(8, genpass.CharsetLower+genpass.CharsetNum)
	return fmt.Sprintf("window-%s", id)
}

func (s *Style) UnmarshalKDL(node *kdl.Node) error {
	if len(node.Arguments) != 0 {
		if len(node.Children) != 0 {
			return errors.New("cannot combine style string and map")
		}
		if str, ok := node.Arguments[0].(kdl.String); ok {
			s.String = fmt.Sprint(str.Value())
		} else {
			return fmt.Errorf("invalid style string: %v", node.Arguments[0])
		}
		return nil
	}

	s.Map = make(map[string]string)
	for _, child := range node.Children {
		key := child.Name
		if len(child.Arguments) == 0 {
			return fmt.Errorf("style map entry %s has no value", key)
		}

		var value strings.Builder

		if len(child.Arguments) == 1 {
			if _, ok := child.Arguments[0].(kdl.Integer); ok {
				if i, err := strconv.Atoi(fmt.Sprint(child.Arguments[0].Value())); err == nil {
					if key != "font-weight" {
						s.Map[key] = fmt.Sprintf("%dpx", i)
						continue
					}
				}
			}
		}

		for i, arg := range child.Arguments {
			if i > 0 {
				value.WriteString(" ")
			}
			value.WriteString(fmt.Sprint(arg.Value()))
		}
		s.Map[key] = value.String()
	}

	return nil
}

func (p *Position) UnmarshalKDL(node *kdl.Node) error {
	if top, ok := node.Properties["top"]; ok {
		if i, err := strconv.Atoi(fmt.Sprint(top.Value())); err == nil {
			p.Top = &i
		} else {
			return fmt.Errorf("invalid top value: %v", top)
		}
	}
	if bottom, ok := node.Properties["bottom"]; ok {
		if i, err := strconv.Atoi(fmt.Sprint(bottom.Value())); err == nil {
			p.Bottom = &i
		} else {
			return fmt.Errorf("invalid bottom value: %v", bottom)
		}
	}
	if left, ok := node.Properties["left"]; ok {
		if i, err := strconv.Atoi(fmt.Sprint(left.Value())); err == nil {
			p.Left = &i
		} else {
			return fmt.Errorf("invalid left value: %v", left)
		}
	}
	if right, ok := node.Properties["right"]; ok {
		if i, err := strconv.Atoi(fmt.Sprint(right.Value())); err == nil {
			p.Right = &i
		} else {
			return fmt.Errorf("invalid right value: %v", right)
		}
	}
	if len(node.Arguments) > 0 && fmt.Sprint(node.Arguments[0].Value()) == "center" {
		p.Center = true
	}

	if len(node.Arguments) > 1 {
		return fmt.Errorf("too many arguments for position: %v", node.Arguments)
	}

	return nil
}

var units = map[string]time.Duration{
	"hours":        time.Hour,
	"hour":         time.Hour,
	"hr":           time.Hour,
	"h":            time.Hour,
	"minutes":      time.Minute,
	"minute":       time.Minute,
	"min":          time.Minute,
	"m":            time.Minute,
	"seconds":      time.Second,
	"second":       time.Second,
	"sec":          time.Second,
	"s":            time.Second,
	"milliseconds": time.Millisecond,
	"millisecond":  time.Millisecond,
	"ms":           time.Millisecond,
}

func (t *TimeSpec) UnmarshalKDL(node *kdl.Node) error {
	parts := node.Arguments
	if len(parts)%2 != 0 {
		return fmt.Errorf("invalid time spec: %#v", node)
	}

	var total time.Duration
	for i := 0; i < len(parts); i += 2 {
		amountValue := parts[i]
		var amount float64
		var err error
		switch v := amountValue.(type) {
		case kdl.Float:
			amount, err = strconv.ParseFloat(fmt.Sprint(v.Value()), 64)
		case kdl.Integer:
			var amountInt int64
			amountInt, err = strconv.ParseInt(fmt.Sprint(v.Value()), 10, 64)
			amount = float64(amountInt)
		default:
			return fmt.Errorf("invalid time spec value type: %#v in %#v", v, node)
		}
		if err != nil {
			return fmt.Errorf("invalid time spec value: in %#v; %w", node, err)
		}
		if unit, ok := parts[i+1].(kdl.String); !ok {
			return fmt.Errorf("invalid time spec unit: %#v in %#v", parts[i+1], node)
		} else {
			unitDuration, ok := units[fmt.Sprint(unit.Value())]
			if !ok {
				return fmt.Errorf("invalid time spec unit: %s in %#v", unit.Value(), node)
			}
			total += time.Duration(amount * float64(unitDuration))
		}
	}

	*t = TimeSpec(total)
	return nil
}
