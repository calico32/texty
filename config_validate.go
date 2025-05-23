package texty

import (
	"fmt"
	"os"
)

func (c Config) Validate() error {
	if len(c.Windows) == 0 {
		return fmt.Errorf("no windows defined")
	}

	for i, window := range c.Windows {
		if window.Id == "" {
			return fmt.Errorf("window #%d: id is required", i)
		}
		textSourceCount := 0
		if window.Command != nil && len(window.Command) > 0 {
			textSourceCount++
		}
		if window.Text != nil && *window.Text != "" {
			textSourceCount++
		}
		if window.File != nil && *window.File != "" {
			textSourceCount++
		}
		if textSourceCount == 0 {
			return fmt.Errorf("window #%d: one of command, text, or file is required", i)
		}
		if textSourceCount > 1 {
			return fmt.Errorf("window #%d: only one of command, text, or file is allowed", i)
		}
		if window.Interval != nil {
			// not valid with text
			if window.Text != nil && *window.Text != "" {
				return fmt.Errorf("window #%d: interval is not valid with text", i)
			}
		}

	}

	if c.Defaults.Style != nil && c.Defaults.Style.String != "" {
		return fmt.Errorf("defaults: style cannot be a string (use map instead)")
	}

	if c.Styles != "" {
		// validate path
		if _, err := os.Stat(c.Styles); os.IsNotExist(err) {
			return fmt.Errorf("styles: CSS file does not exist: %s", c.Styles)
		}
	}

	return nil
}
