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
	astilectronPath   = flag.String("a", "", "the astilectron path")
	configurationPath = flag.String("c", "", "the configuration path")
	darwin            = flag.Bool("d", false, "if set, will add darwin/amd64 to the environments")
	linux             = flag.Bool("l", false, "if set, will add linux/amd64 to the environments")
	windows           = flag.Bool("w", false, "if set, will add windows/amd64 to the environments")
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

	// Astilectron path
	if len(*astilectronPath) > 0 {
		c.AstilectronPath = *astilectronPath
	}

	// Environments
	if *darwin {
		c.Environments = append(c.Environments, astibundler.ConfigurationEnvironment{Arch: "amd64", OS: "darwin"})
	}
	if *linux {
		c.Environments = append(c.Environments, astibundler.ConfigurationEnvironment{Arch: "amd64", OS: "linux"})
	}
	if *windows {
		c.Environments = append(c.Environments, astibundler.ConfigurationEnvironment{Arch: "amd64", OS: "windows"})
	}
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
		if err = b.ClearCache(); err != nil {
			astilog.Fatal(errors.Wrap(err, "clearing cache failed"))
		}
	default:
		// Bundle
		if err = b.Bundle(); err != nil {
			astilog.Fatal(errors.Wrap(err, "bundling failed"))
		}
	}
}
