package astibundler

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/os"
	"github.com/asticode/go-astitools/slice"
	"github.com/jteeuwen/go-bindata"
	"github.com/pkg/errors"
)

// Bundler represents an object capable of bundling an Astilectron app
type Bundler struct {
	appName            string
	environments       []ConfigurationEnvironment
	pathAppIconDarwin  string
	pathAppIconDefault string
	pathBuild          string
	pathInput          string
	pathOutput         string
}

// New builds a new bundler based on a configuration
func New(c *Configuration) (b *Bundler, err error) {
	// Init
	b = &Bundler{
		appName:      c.AppName,
		environments: c.Environments,
	}

	// Darwin app icon path
	if len(c.AppIconDarwinPath) > 0 {
		if b.pathAppIconDarwin, err = filepath.Abs(c.AppIconDarwinPath); err != nil {
			err = errors.Wrapf(err, "filepath.Abs of %s failed", c.AppIconDarwinPath)
			return
		}
	}

	// Default app icon path
	if len(c.AppIconDefaultPath) > 0 {
		if b.pathAppIconDefault, err = filepath.Abs(c.AppIconDefaultPath); err != nil {
			err = errors.Wrapf(err, "filepath.Abs of %s failed", c.AppIconDefaultPath)
			return
		}
	}

	// Input path
	if len(c.InputPath) > 0 {
		if b.pathInput, err = filepath.Abs(c.InputPath); err != nil {
			err = errors.Wrapf(err, "filepath.Abs of %s failed", c.InputPath)
			return
		}
	} else {
		// Default to current directory
		if b.pathInput, err = os.Getwd(); err != nil {
			err = errors.Wrap(err, "os.Getwd failed")
			return
		}
	}

	// Build path
	b.pathBuild = strings.TrimPrefix(strings.TrimPrefix(b.pathInput, filepath.Join(os.Getenv("GOPATH"), "src")), string(os.PathSeparator))

	// Output path
	if len(c.OutputPath) > 0 {
		if b.pathOutput, err = filepath.Abs(c.OutputPath); err != nil {
			err = errors.Wrapf(err, "filepath.Abs of %s failed", c.OutputPath)
			return
		}
	} else {
		// Default to current directory
		if b.pathOutput, err = os.Getwd(); err != nil {
			err = errors.Wrap(err, "os.Getwd failed")
			return
		}
	}
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
	for _, e := range b.environments {
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
	astilog.Debugf("Creating %s", b.pathOutput)
	if err = os.MkdirAll(b.pathOutput, 0777); err != nil {
		err = errors.Wrapf(err, "mkdirall %s failed", b.pathOutput)
		return
	}
	return
}

// bindResources binds the resources
func (b *Bundler) bindResources() (err error) {
	// Init paths
	var ip = filepath.Join(b.pathInput, "resources")
	var op = filepath.Join(b.pathInput, "resources.go")

	// No resources folder
	if i, errStat := os.Stat(ip); os.IsNotExist(errStat) || !i.IsDir() {
		return
	}

	// Init bindata config
	var c = bindata.NewConfig()
	c.Input = []bindata.InputConfig{{Path: ip, Recursive: true}}
	c.Output = op
	c.Prefix = b.pathInput

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
	var environmentPath = filepath.Join(b.pathOutput, e.OS, e.Arch)
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
	var cmd = exec.Command("go", "build", "-ldflags", `-X "main.AppName=`+b.appName+`" -X "main.BuiltAt=`+time.Now().String()+`"`, "-o", binaryPath, b.pathBuild)
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
	var contentsPath = filepath.Join(environmentPath, b.appName+".app", "Contents")
	var macOSPath = filepath.Join(contentsPath, "MacOS")
	astilog.Debugf("Creating %s", macOSPath)
	if err = os.MkdirAll(macOSPath, 0777); err != nil {
		err = errors.Wrapf(err, "mkdirall of %s failed", macOSPath)
		return
	}

	// Move binary
	var macOSBinaryPath = filepath.Join(macOSPath, b.appName)
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
	if len(b.pathAppIconDarwin) > 0 {
		// Create Resources folder
		var resourcesPath = filepath.Join(contentsPath, "Resources")
		astilog.Debugf("Creating %s", resourcesPath)
		if err = os.MkdirAll(resourcesPath, 0777); err != nil {
			err = errors.Wrapf(err, "mkdirall of %s failed", resourcesPath)
			return
		}

		// Copy icon
		var ip = filepath.Join(resourcesPath, b.appName+filepath.Ext(b.pathAppIconDarwin))
		astilog.Debugf("Copying %s to %s", b.pathAppIconDarwin, ip)
		if err = astios.Copy(context.Background(), b.pathAppIconDarwin, ip); err != nil {
			err = errors.Wrapf(err, "copying %s to %s failed", b.pathAppIconDarwin, ip)
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
		<string>`+b.appName+filepath.Ext(b.pathAppIconDarwin)+`</string>
		<key>CFBundleDisplayName</key>
		<string>`+b.appName+`</string>
		<key>CFBundleExecutable</key>
		<string>`+b.appName+`</string>
		<key>CFBundleName</key>
		<string>`+b.appName+`</string>
		<key>CFBundleIdentifier</key>
		<string>com.`+b.appName+`</string>
	</dict>
</plist>`), 0777); err != nil {
		err = errors.Wrapf(err, "adding Info.plist to %s failed", fp)
		return
	}
	return
}
