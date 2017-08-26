package astilectron_bundler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/os"
	"github.com/asticode/go-astitools/slice"
	"github.com/jteeuwen/go-bindata"
	"github.com/pkg/errors"
)

// Bundler represents an object capable of bundling an Astilectron app
type Bundler struct {
	buildPath string
	c         *Configuration
}

// New builds a new bundler based on a configuration path
func New(configurationPath string) (b *Bundler, err error) {
	// Open file
	var f *os.File
	if f, err = os.Open(configurationPath); err != nil {
		err = errors.Wrapf(err, "opening file %s failed", configurationPath)
		return
	}
	defer f.Close()

	// Unmarshal
	b = &Bundler{c: &Configuration{}}
	if err = json.NewDecoder(f).Decode(b.c); err != nil {
		err = errors.Wrap(err, "unmarshaling configuration failed")
		return
	}

	// Add build path
	if b.buildPath, err = buildPath(b.c.InputPath); err != nil {
		err = errors.Wrap(err, "building go path failed")
		return
	}
	return
}

// buildPath builds a build paths i.e. github.com/asticode/go-project for /path/to/github.com/asticode/go-project
func buildPath(i string) (o string, err error) {
	var ps []string
	var p = i
	for len(ps) < 3 {
		var b = filepath.Base(p)
		if b == string(filepath.Separator) {
			break
		}
		ps = append([]string{b}, ps...)
		p = filepath.Dir(p)
	}
	if len(ps) < 3 {
		err = fmt.Errorf("couldn't parse build path of %s", i)
		return
	}
	o = filepath.Join(ps...)
	return
}

// Bundle bundles an astilectron app based on a configuration
func (b *Bundler) Bundle() (err error) {
	// Reset
	astilog.Debug("Resetting")
	if err = b.reset(); err != nil {
		err = errors.Wrap(err, "resetting bundler failed")
		return
	}

	// Bind resources
	astilog.Debug("Binding resources")
	if err = b.bindResources(); err != nil {
		err = errors.Wrap(err, "binding resources failed")
		return
	}

	// Loop through environments
	for _, e := range b.c.Environments {
		astilog.Debugf("Bundling for environment %s/%s", e.OS, e.Arch)
		if err = b.bundle(e); err != nil {
			err = errors.Wrapf(err, "bundling for environment %s/%s failed", e.OS, e.Arch)
			return
		}
	}
	return
}

// reset resets the bundler
func (b *Bundler) reset() (err error) {
	// Make sure the output path exists
	astilog.Debugf("Creating %s", b.c.OutputPath)
	if err = os.MkdirAll(b.c.OutputPath, 0777); err != nil {
		err = errors.Wrapf(err, "mkdirall %s failed", b.c.OutputPath)
		return
	}
	return
}

// bindResources binds the resources
func (b *Bundler) bindResources() (err error) {
	// Init paths
	var ip = filepath.Join(b.c.InputPath, "resources")
	var op = filepath.Join(b.c.InputPath, "resources.go")

	// No resources folder
	if i, errStat := os.Stat(ip); os.IsNotExist(errStat) || !i.IsDir() {
		return
	}

	// Init bindata config
	var c = bindata.NewConfig()
	c.Input = []bindata.InputConfig{{Path: ip, Recursive: true}}
	c.Output = op
	c.Prefix = b.c.InputPath

	// Bind data
	astilog.Debugf("Binding %s into %s", ip, op)
	err = bindata.Translate(c)
	return
}

// bundle bundles an os
func (b *Bundler) bundle(e ConfigurationEnvironment) (err error) {
	// Validate OS
	if !astislice.InStringSlice(e.OS, astilectron.ValidOSes()) {
		err = fmt.Errorf("OS %s is not supported", e.OS)
		return
	}

	// Remove previous environment folder
	var environmentPath = filepath.Join(b.c.OutputPath, e.OS, e.Arch)
	astilog.Debugf("Removing %s", environmentPath)
	if err = os.RemoveAll(environmentPath); err != nil {
		err = errors.Wrapf(err, "removing %s failed", environmentPath)
		return
	}

	// Create the environment folder
	astilog.Debugf("Creating %s", environmentPath)
	if err = os.MkdirAll(environmentPath, 0777); err != nil {
		err = errors.Wrapf(err, "mkdirall %s failed", environmentPath)
		return
	}

	// TODO Bind astilectron data

	// Build
	astilog.Debug("Building")
	var binaryPath = filepath.Join(environmentPath, "binary")
	var cmd = exec.Command("go", "build", "-o", binaryPath, b.buildPath)
	cmd.Env = []string{
		"GOARCH=" + e.Arch,
		"GOOS=" + e.OS,
		"GOPATH=" + os.Getenv("GOPATH"),
	}
	var o []byte
	if o, err = cmd.CombinedOutput(); err != nil {
		err = errors.Wrapf(err, "building failed: %s", o)
		return
	}

	// Finish bundle based on OS
	switch e.OS {
	case "darwin":
		err = b.finishDarwin(environmentPath, binaryPath)
	default:
		err = fmt.Errorf("OS %s is not yet implemented", e.OS)
	}
	return
}

// finishDarwin finishes bundle for a darwin system
func (b *Bundler) finishDarwin(environmentPath, binaryPath string) (err error) {
	// Create MacOS folder
	var contentsPath = filepath.Join(environmentPath, b.c.AppName+".app", "Contents")
	var macOSPath = filepath.Join(contentsPath, "MacOS")
	astilog.Debugf("Creating %s", macOSPath)
	if err = os.MkdirAll(macOSPath, 0777); err != nil {
		err = errors.Wrapf(err, "mkdirall of %s failed", macOSPath)
		return
	}

	// Move binary
	var macOSBinaryPath = filepath.Join(macOSPath, b.c.AppName)
	astilog.Debugf("Moving %s to %s", binaryPath, macOSBinaryPath)
	if err = astios.Move(context.Background(), binaryPath, macOSBinaryPath); err != nil {
		err = errors.Wrapf(err, "moving %s to %s failed", binaryPath, macOSBinaryPath)
		return
	}

	// Make sure the binary is executable
	astilog.Debugf("Chmoding %s", macOSBinaryPath)
	if err = os.Chmod(macOSBinaryPath, 0777); err != nil {
		err = errors.Wrapf(err, "chmoding %s failed", macOSBinaryPath)
		return
	}

	// App icon
	if len(b.c.AppIconDarwinPath) > 0 {
		// Create Resources folder
		var resourcesPath = filepath.Join(contentsPath, "Resources")
		astilog.Debugf("Creating %s", resourcesPath)
		if err = os.MkdirAll(resourcesPath, 0777); err != nil {
			err = errors.Wrapf(err, "mkdirall of %s failed", resourcesPath)
			return
		}

		// Copy icon
		var ip = filepath.Join(resourcesPath, b.c.AppName+filepath.Ext(b.c.AppIconDarwinPath))
		astilog.Debugf("Copying %s to %s", b.c.AppIconDarwinPath, ip)
		if err = astios.Copy(context.Background(), b.c.AppIconDarwinPath, ip); err != nil {
			err = errors.Wrapf(err, "copying %s to %s failed", b.c.AppIconDarwinPath, ip)
			return
		}
	}

	// Add Info.plist file
	var fp = filepath.Join(contentsPath, "Info.plist")
	astilog.Debugf("Adding Info.plist to %s", fp)
	if err = ioutil.WriteFile(fp, []byte(`<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>CFBundleIconFile</key>
		<string>`+b.c.AppName+filepath.Ext(b.c.AppIconDarwinPath)+`</string>
		<key>CFBundleDisplayName</key>
		<string>`+b.c.AppName+`</string>
		<key>CFBundleExecutable</key>
		<string>`+b.c.AppName+`</string>
		<key>CFBundleName</key>
		<string>`+b.c.AppName+`</string>
		<key>CFBundleIdentifier</key>
		<string>com.`+b.c.AppName+`</string>
	</dict>
</plist>`), 0777); err != nil {
		err = errors.Wrapf(err, "adding Info.plist to %s failed", fp)
		return
	}
	return
}
