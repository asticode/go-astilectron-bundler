package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/asticode/go-astilectron-bundler"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// Flags
var (
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

	// Bundle
	if err = astibundler.New(c).Bundle(); err != nil {
		astilog.Fatal(errors.Wrap(err, "bundling failed"))
	}
}
