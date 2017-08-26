package main

import (
	"encoding/json"
	"flag"
	"os"
	"runtime"

	"github.com/asticode/go-astilectron-bundler"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// Flags
var (
	autoEnvironment   = flag.Bool("a", false, "if set, the bundler environment is the current one")
	configurationPath = flag.String("c", "", "the configuration path")
)

func main() {
	// Init
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

	// Auto environment
	if *autoEnvironment {
		c.Environments = []astibundler.ConfigurationEnvironment{{Arch: runtime.GOARCH, OS: runtime.GOOS}}
	}

	// Build bundler
	var b *astibundler.Bundler
	if b, err = astibundler.New(c); err != nil {
		astilog.Fatal(errors.Wrap(err, "building bundler failed"))
	}

	// Bundle
	if err = b.Bundle(); err != nil {
		astilog.Fatal(errors.Wrap(err, "bundling failed"))
	}
}
