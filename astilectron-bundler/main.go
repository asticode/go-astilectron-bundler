package main

import (
	"flag"

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

	// Build bundler
	var b *astibundler.Bundler
	var err error
	if b, err = astibundler.New(*configurationPath); err != nil {
		astilog.Fatal(errors.Wrapf(err, "new bundler for configuration path %s failed", *configurationPath))
	}

	// Bundle
	if err = b.Bundle(); err != nil {
		astilog.Fatal(errors.Wrap(err, "bundling failed"))
	}
}
