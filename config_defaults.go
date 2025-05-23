package texty

import (
	layershell "github.com/diamondburned/gotk-layer-shell"
)

func (c *Config) ApplyDefaults() {
	for _, window := range c.Windows {
		hasNoSources := window.Command == nil && window.Text == nil && window.File == nil
		if c.Defaults.Command != nil && hasNoSources {
			window.Command = c.Defaults.Command
		}
		if c.Defaults.Text != nil && hasNoSources {
			window.Text = c.Defaults.Text
		}
		if c.Defaults.File != nil && hasNoSources {
			window.File = c.Defaults.File
		}

		if c.Defaults.Interval != nil && window.Interval == nil && window.Text == nil {
			// only apply interval if there is no text
			window.Interval = c.Defaults.Interval
		}

		if c.Defaults.Position != nil && window.Position == nil {
			window.Position = c.Defaults.Position
		}
		if c.Defaults.Layer != nil && window.Layer == nil {
			window.Layer = c.Defaults.Layer
		} else if window.Layer == nil {
			// layer must be set, default to bottom
			layer := layershell.LayerBottom
			window.Layer = &layer
		}

		if c.Defaults.Style != nil {
			if window.Style == nil {
				// simply copy the style
				window.Style = c.Defaults.Style
			} else if window.Style.String != "" {
				// we can't apply defaults to a string
				// skip
			} else {
				// merge maps
				for k, v := range c.Defaults.Style.Map {
					if _, ok := window.Style.Map[k]; !ok {
						// only add if not already set
						window.Style.Map[k] = v
					}
				}
			}
		}
	}
}
