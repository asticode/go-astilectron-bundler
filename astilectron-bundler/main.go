package main

import (
	"encoding/json"
	"flag"
	"os"
	"runtime"

	"github.com/asticode/go-astilectron-bundler"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/flag"
	"github.com/pkg/errors"
)

// Flags
var (
	configurationPath = flag.String("c", "", "the configuration path")
)

func main() {
	// Init
	var s = astiflag.Subcommand()
	flag.Parse()
	astilog.FlagInit()

	// Open file
	var f *os.File
	var err error
	if f, err = os.Open(*configurationPath); err != nil {
		astilog.Fatal(errors.Wrapf(err, "opening file %s failed", *configurationPath))
	}
	defer f.Close()

	// Unmarshal
	var c *astibundler.Configuration
	if err = json.NewDecoder(f).Decode(&c); err != nil {
		astilog.Fatal(errors.Wrap(err, "unmarshaling configuration failed"))
	}

	// Default environment
	if len(c.Environments) == 0 {
		c.Environments = []astibundler.ConfigurationEnvironment{{Arch: runtime.GOARCH, OS: runtime.GOOS}}
	}

	// Build bundler
	var b *astibundler.Bundler
	if b, err = astibundler.New(c); err != nil {
		astilog.Fatal(errors.Wrap(err, "building bundler failed"))
	}

	// Switch on subcommand
	switch s {
	case "cc":
		// Clear cache

	default:
		// Bundle
		if err = b.Bundle(); err != nil {
			astilog.Fatal(errors.Wrap(err, "bundling failed"))
		}
	}
}
