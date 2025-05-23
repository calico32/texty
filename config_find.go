package texty

import (
	"errors"
	"log"
	"os"

	"github.com/calico32/kdl-go"
	layershell "github.com/diamondburned/gotk-layer-shell"
)

var defaultText = "Welcome to texty! This is an example window that appears when you don't have a config file.\n\nTo get started, write your configuration to ~/.config/texty/config.kdl.\nFor documentation, visit https://github.com/calico32/texty.\n\nMiddle click on this window to stop texty."
var defaultLayer = layershell.LayerTop
var defaultMargin = 16
var DefaultConfig = Config{
	Windows: []*Window{
		{
			Id:    generateRandomId(),
			Text:  &defaultText,
			Layer: &defaultLayer,
			Position: &Position{
				Top:  &defaultMargin,
				Left: &defaultMargin,
			},
			Style: &Style{
				Map: map[string]string{
					"background-color": "#104e64",
					"color":            "#a2f4fd",
					"font-size":        "16px",
					"border-width":     "1px",
					"border-color":     "#007595",
					"border-style":     "solid",
					"border-radius":    "4px",
				},
			},
		},
	},
}

func MakeErrorConfig(err error) Config {
	text := "texty failed to start!\nError loading config: " + err.Error()
	layer := layershell.LayerTop
	margin := 16
	return Config{
		Windows: []*Window{
			{
				Id:    generateRandomId(),
				Text:  &text,
				Layer: &layer,
				Position: &Position{
					Top:  &margin,
					Left: &margin,
				},
				Style: &Style{
					Map: map[string]string{
						"background-color": "#82181a",
						"color":            "#ffe2e2",
						"font-size":        "16px",
						"border-width":     "1px",
						"border-color":     "#c10007",
						"border-style":     "solid",
						"border-radius":    "4px",
					},
				},
			},
		},
	}
}

var configPaths []string

func init() {
	paths := []string{}

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		paths = append(paths, xdgConfigHome+"/texty/config.kdl")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	paths = append(paths, homeDir+"/.config/texty/config.kdl")
	paths = append(paths, homeDir+"/.texty/config.kdl")

	configPaths = paths
}

func LoadConfig(customPath *string, verbose bool) (Config, string, error) {
	if customPath != nil && *customPath != "" {
		config, err, _ := tryConfig(*customPath, verbose)
		return config, *customPath, err
	}
	for _, path := range configPaths {
		config, err, shouldReturn := tryConfig(path, verbose)
		if shouldReturn {
			return config, path, err
		}
	}

	return DefaultConfig, "", errors.New("no config file found")
}

func tryConfig(path string, verbose bool) (Config, error, bool) {
	if verbose {
		log.Printf("trying config file: %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		if verbose {
			log.Printf("failed to open config file: %s", path)
		}
		// LoadConfig should not return if the file does not exist; it should
		// try the next path
		return MakeErrorConfig(err), err, !os.IsNotExist(err)
	}
	defer f.Close()
	doc, err := kdl.NewParser(kdl.KdlVersion2, f).ParseDocument()
	if err != nil {
		return MakeErrorConfig(err), err, true
	}
	var config Config
	if err := config.UnmarshalKDL(doc); err != nil {
		return MakeErrorConfig(err), err, true
	}

	if verbose {
		log.Printf("loaded config file:\n%s\n", config.SerializeJSON())
	}

	config.ApplyDefaults()

	if verbose {
		log.Printf("applied defaults:\n%s\n", config.SerializeJSON())
	}

	if err := config.Validate(); err != nil {
		return MakeErrorConfig(err), err, true
	}

	return config, nil, true
}
