package texty_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"texty"

	"github.com/calico32/kdl-go"
)

func TestConfig(t *testing.T) {
	f, err := os.Open("config.kdl")
	if err != nil {
		t.Fatalf("failed to open config file: %v", err)
	}
	defer f.Close()

	doc, err := kdl.NewParser(kdl.KdlVersion2, f).ParseDocument()
	if err != nil {
		t.Fatalf("failed to decode config: %v", err)
	}

	var config texty.Config
	if err := config.UnmarshalKDL(doc); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	fmt.Println(string(out))
}
