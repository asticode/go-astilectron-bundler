package astibundler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/os"
	"github.com/asticode/rsrc"
	"github.com/jteeuwen/go-bindata"
	"github.com/pkg/errors"
)

// Configuration represents the bundle configuration
type Configuration struct {
	// The app name as it should be displayed everywhere
	// It's also set as an ldflag and therefore accessible in a global var main.AppName
	AppName string `json:"app_name"`

	// The bundler cache the vendor content in this path.
	// Best is to leave it empty.
	CachePath string `json:"cache_path"`

	// List of environments the bundling should be done upon.
	// An environment is a combination of OS and ARCH
	Environments []ConfigurationEnvironment `json:"environments"`

	// Paths to icons
	IconPathDarwin  string `json:"icon_path_darwin"` // .icns
	IconPathLinux   string `json:"icon_path_linux"`
	IconPathWindows string `json:"icon_path_windows"` // .ico

	// The path of the project.
	// Best is to leave it empty and execute the bundler while in the project folder
	InputPath string `json:"input_path"`

	// The path where the files will be written
	OutputPath string `json:"output_path"`
}

// ConfigurationEnvironment represents the bundle configuration environment
type ConfigurationEnvironment struct {
	Arch string `json:"arch"`
	OS   string `json:"os"`
}

// Bundler represents an object capable of bundling an Astilectron app
type Bundler struct {
	appName         string
	Client          *http.Client
	environments    []ConfigurationEnvironment
	pathBuild       string
	pathCache       string
	pathIconDarwin  string
	pathIconLinux   string
	pathIconWindows string
	pathInput       string
	pathOutput      string
	pathResources   string
	pathVendor      string
}

// New builds a new bundler based on a configuration
func New(c *Configuration) (b *Bundler, err error) {
	// Init
	b = &Bundler{
		appName:      c.AppName,
		Client:       &http.Client{},
		environments: c.Environments,
	}

	// Loop through environments
	for _, env := range b.environments {
		// Validate OS
		if !astilectron.IsValidOS(env.OS) {
			err = fmt.Errorf("OS %s is invalid", env.OS)
			return
		}
	}

	// Cache path
	if len(c.CachePath) > 0 {
		if b.pathCache, err = filepath.Abs(c.CachePath); err != nil {
			err = errors.Wrapf(err, "filepath.Abs of %s failed", c.CachePath)
			return
		}
	} else {
		b.pathCache = filepath.Join(os.TempDir(), "astibundler")
	}

	// Darwin icon path
	if len(c.IconPathDarwin) > 0 {
		if b.pathIconDarwin, err = filepath.Abs(c.IconPathDarwin); err != nil {
			err = errors.Wrapf(err, "filepath.Abs of %s failed", c.IconPathDarwin)
			return
		}
	}

	// Linux icon path
	if len(c.IconPathLinux) > 0 {
		if b.pathIconLinux, err = filepath.Abs(c.IconPathLinux); err != nil {
			err = errors.Wrapf(err, "filepath.Abs of %s failed", c.IconPathLinux)
			return
		}
	}

	// Windows icon path
	if len(c.IconPathWindows) > 0 {
		if b.pathIconWindows, err = filepath.Abs(c.IconPathWindows); err != nil {
			err = errors.Wrapf(err, "filepath.Abs of %s failed", c.IconPathWindows)
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

	// Paths that depends on the input path
	b.pathBuild = strings.TrimPrefix(strings.TrimPrefix(b.pathInput, filepath.Join(os.Getenv("GOPATH"), "src")), string(os.PathSeparator))
	b.pathResources = filepath.Join(b.pathInput, "resources")
	b.pathVendor = filepath.Join(b.pathInput, "vendor")

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

// ClearCache clears the bundler cache
func (b *Bundler) ClearCache() (err error) {
	// Remove cache folder
	astilog.Debugf("Removing %s", b.pathCache)
	if err = os.RemoveAll(b.pathCache); err != nil {
		err = errors.Wrapf(err, "removing %s failed", b.pathCache)
		return
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
	// Make sure the minimal paths exist
	for _, path := range []string{b.pathCache, b.pathOutput} {
		astilog.Debugf("Creating %s", path)
		if err = os.MkdirAll(path, 0777); err != nil {
			err = errors.Wrapf(err, "mkdirall %s failed", path)
			return
		}
	}
	return
}

// provisionVendorZip provisions a vendor zip file
func (b *Bundler) provisionVendorZip(pathDownload, pathCache, pathVendor string) (err error) {
	// Download source
	if _, errStat := os.Stat(pathCache); os.IsNotExist(errStat) {
		if err = astilectron.Download(context.Background(), b.Client, pathDownload, pathCache); err != nil {
			err = errors.Wrapf(err, "downloading %s into %s failed", pathDownload, pathCache)
			return
		}
	} else {
		astilog.Debugf("%s already exists, skipping download of %s", pathCache, pathDownload)
	}

	// Copy
	astilog.Debugf("Copying %s to %s", pathCache, pathVendor)
	if err = astios.Copy(context.Background(), pathCache, pathVendor); err != nil {
		err = errors.Wrapf(err, "copying %s to %s failed", pathCache, pathVendor)
	}
	return
}

// provisionVendorAstilectron provisions the astilectron vendor zip file
func (b *Bundler) provisionVendorAstilectron() error {
	return b.provisionVendorZip(astilectron.AstilectronDownloadSrc(), filepath.Join(b.pathCache, fmt.Sprintf("astilectron-%s.zip", astilectron.VersionAstilectron)), filepath.Join(b.pathVendor, zipNameAstilectron))
}

// provisionVendorElectron provisions the electron vendor zip file
func (b *Bundler) provisionVendorElectron(oS, arch string) error {
	return b.provisionVendorZip(astilectron.ElectronDownloadSrc(oS, arch), filepath.Join(b.pathCache, fmt.Sprintf("electron-%s-%s-%s.zip", oS, arch, astilectron.VersionElectron)), filepath.Join(b.pathVendor, zipNameElectron))
}

// provisionVendor provisions the vendor folder
func (b *Bundler) provisionVendor(oS, arch string) (err error) {
	// Remove previous vendor folder
	astilog.Debugf("Removing %s", b.pathVendor)
	if err = os.RemoveAll(b.pathVendor); err != nil {
		err = errors.Wrapf(err, "removing %s failed", b.pathVendor)
		return
	}

	// Create the vendor folder
	astilog.Debugf("Creating %s", b.pathVendor)
	if err = os.MkdirAll(b.pathVendor, 0777); err != nil {
		err = errors.Wrapf(err, "mkdirall %s failed", b.pathVendor)
		return
	}

	// Provision astilectron
	if err = b.provisionVendorAstilectron(); err != nil {
		err = errors.Wrap(err, "provisioning astilectron vendor failed")
		return
	}

	// Provision electron
	if err = b.provisionVendorElectron(oS, arch); err != nil {
		err = errors.Wrapf(err, "provisioning electron vendor for OS %s and arch %s failed", oS, arch)
		return
	}
	return
}

// bindData binds the data
func (b *Bundler) bindData(os, arch string) (err error) {
	// Provision the vendor
	if err = b.provisionVendor(os, arch); err != nil {
		err = errors.Wrap(err, "provisioning the vendor failed")
		return
	}

	// Build bindata config
	var c = bindata.NewConfig()
	c.Input = []bindata.InputConfig{
		{Path: filepath.Join(b.pathInput, "resources"), Recursive: true},
		{Path: filepath.Join(b.pathInput, "vendor"), Recursive: true},
	}
	c.Output = filepath.Join(b.pathInput, "bind.go")
	c.Prefix = b.pathInput

	// Bind data
	err = bindata.Translate(c)
	return
}

// addWindowsSyso adds the proper windows .syso if needed
func (b *Bundler) addWindowsSyso(arch string) (err error) {
	if len(b.pathIconWindows) > 0 {
		var p = filepath.Join(b.pathInput, "windows.syso")
		astilog.Debugf("Running rsrc for icon %s into %s", b.pathIconWindows, p)
		if err = rsrc.Run("", b.pathIconWindows, p, arch); err != nil {
			err = errors.Wrapf(err, "running rsrc for icon %s into %s failed", b.pathIconWindows, p)
			return
		}
	}
	return
}

// ldflags represents ldflags
type ldflags map[string][]string

// string returns the ldflags as a string
func (l ldflags) string() string {
	var o []string
	for k, ss := range l {
		for _, s := range ss {
			o = append(o, fmt.Sprintf(`-%s %s`, k, s))
		}
	}
	return strings.Join(o, " ")
}

// bundle bundles an os
func (b *Bundler) bundle(e ConfigurationEnvironment) (err error) {
	// Remove previous environment folder
	var environmentPath = filepath.Join(b.pathOutput, e.OS+"-"+e.Arch)
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

	// Bind data
	astilog.Debug("Binding data")
	if err = b.bindData(e.OS, e.Arch); err != nil {
		err = errors.Wrap(err, "binding data failed")
		return
	}

	// Add windows .syso
	if e.OS == "windows" {
		if err = b.addWindowsSyso(e.Arch); err != nil {
			err = errors.Wrap(err, "adding windows .syso failed")
			return
		}
	}

	// Build ldflags
	var l = ldflags{
		"X": []string{
			`"main.AppName=` + b.appName + `"`,
			`"main.BuiltAt=` + time.Now().String() + `"`,
		},
	}
	if e.OS == "windows" {
		l["H"] = []string{"windowsgui"}
	}

	// Build cmd
	astilog.Debugf("Building for os %s and arch %s", e.OS, e.Arch)
	var binaryPath = filepath.Join(environmentPath, "binary")
	var cmd = exec.Command("go", "build", "-ldflags", l.string(), "-o", binaryPath, b.pathBuild)
	cmd.Env = []string{
		"GOARCH=" + e.Arch,
		"GOOS=" + e.OS,
		"GOPATH=" + os.Getenv("GOPATH"),
	}

	// Exec
	var o []byte
	astilog.Debugf("Executing %s", strings.Join(cmd.Args, " "))
	if o, err = cmd.CombinedOutput(); err != nil {
		err = errors.Wrapf(err, "building failed: %s", o)
		return
	}

	// Finish bundle based on OS
	switch e.OS {
	case "darwin":
		err = b.finishDarwin(environmentPath, binaryPath)
	case "linux":
		err = b.finishLinux(environmentPath, binaryPath)
	case "windows":
		err = b.finishWindows(environmentPath, binaryPath)
	default:
		err = fmt.Errorf("OS %s is not yet implemented", e.OS)
	}
	return
}

// finishDarwin finishes bundling for a darwin system
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

	// Icon
	if len(b.pathIconDarwin) > 0 {
		// Create Resources folder
		var resourcesPath = filepath.Join(contentsPath, "Resources")
		astilog.Debugf("Creating %s", resourcesPath)
		if err = os.MkdirAll(resourcesPath, 0777); err != nil {
			err = errors.Wrapf(err, "mkdirall of %s failed", resourcesPath)
			return
		}

		// Copy icon
		var ip = filepath.Join(resourcesPath, b.appName+filepath.Ext(b.pathIconDarwin))
		astilog.Debugf("Copying %s to %s", b.pathIconDarwin, ip)
		if err = astios.Copy(context.Background(), b.pathIconDarwin, ip); err != nil {
			err = errors.Wrapf(err, "copying %s to %s failed", b.pathIconDarwin, ip)
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
		<string>`+b.appName+filepath.Ext(b.pathIconDarwin)+`</string>
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

// finishLinux finishes bundling for a linux system
// TODO Add .desktop file
func (b *Bundler) finishLinux(environmentPath, binaryPath string) (err error) {
	// Move binary
	var linuxBinaryPath = filepath.Join(environmentPath, b.appName)
	astilog.Debugf("Moving %s to %s", binaryPath, linuxBinaryPath)
	if err = astios.Move(context.Background(), binaryPath, linuxBinaryPath); err != nil {
		err = errors.Wrapf(err, "moving %s to %s failed", binaryPath, linuxBinaryPath)
		return
	}
	return
}

// finishWindows finishes bundling for a linux system
func (b *Bundler) finishWindows(environmentPath, binaryPath string) (err error) {
	// Move binary
	var windowsBinaryPath = filepath.Join(environmentPath, b.appName+".exe")
	astilog.Debugf("Moving %s to %s", binaryPath, windowsBinaryPath)
	if err = astios.Move(context.Background(), binaryPath, windowsBinaryPath); err != nil {
		err = errors.Wrapf(err, "moving %s to %s failed", binaryPath, windowsBinaryPath)
		return
	}
	return
}
