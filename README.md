This package provides a way to bundle an [astilectron](https://github.com/asticode/go-astilectron) app using the [bootstrap](https://github.com/asticode/go-astilectron-bootstrap).

Check out the [demo](https://github.com/asticode/go-astilectron-demo) to see a working example.

# Installation

Run the following command:

    $ go get -u github.com/asticode/go-astilectron-bundler/...
    
# Configuration

**astilectron-bundler** uses a configuration file to know what it's supposed to do. Here's an example:

```json
{
  "app_name": "Test",
  "environments": [
    {"arch": "amd64", "os": "darwin"},
    {"arch": "amd64", "os": "linux"},
    {
      "arch": "amd64",
      "os": "windows",
      "env": {
        "CC": "x86_64-w64-mingw32-gcc",
        "CXX": "x86_64-w64-mingw32-g++",
        "CGO_ENABLED": "1"
      }
    }
  ],
  "xgo": {
      "enabled": true,
      "deps": ["https://gmplib.org/download/gmp/gmp-6.0.0a.tar.bz2"]
  },
  "icon_path_darwin": "path/to/icon.icns",
  "icon_path_linux": "path/to/icon.png",
  "icon_path_windows": "path/to/icon.ico",
  "input_path": "path/to/src/github.com/username/project",
  "output_path": "path/to/output/directory"
}
```

Paths can be either relative or absolute but we **strongly** encourage to use relative paths.

If no input path is specified, the working directory path is used.

We **strongly** encourage to leave the input path option empty and execute the **bundler** while in the directory of the project you're bundling.

# Using xgo
To build your app using xgo you should add 
`"xgo": {
      "enabled": true,
      "deps": ["https://gmplib.org/download/gmp/gmp-6.0.0a.tar.bz2"]
  }`
into your `bundler.json`

To install and/or update xgo, simply type:
`go get -u github.com/karalabe/xgo`
You can test whether xgo is functioning correctly by requesting it to cross compile itself and verifying that all cross compilations succeeded or not.
```bash
$ xgo github.com/karalabe/xgo
...

$ ls -al
-rwxr-xr-x  1 root     root      2792436 Sep 14 16:45 xgo-android-21-arm
-rwxr-xr-x  1 root     root      2353212 Sep 14 16:45 xgo-darwin-386
-rwxr-xr-x  1 root     root      2906128 Sep 14 16:45 xgo-darwin-amd64
-rwxr-xr-x  1 root     root      2388288 Sep 14 16:45 xgo-linux-386
-rwxr-xr-x  1 root     root      2960560 Sep 14 16:45 xgo-linux-amd64
-rwxr-xr-x  1 root     root      2437864 Sep 14 16:45 xgo-linux-arm
-rwxr-xr-x  1 root     root      2551808 Sep 14 16:45 xgo-windows-386.exe
-rwxr-xr-x  1 root     root      3130368 Sep 14 16:45 xgo-windows-amd64.exe
```

# Usage

If **astilectron-bundler** has been installed properly (and the $GOPATH is in your $PATH), run the following command:

    $ astilectron-bundler -v -c <path to your configuration file>
    
or if your working directory is your project directory and your bundler configuration has the proper name (`bundler.json`)

    $ astilectron-bundler -v
    
# Output

For each environment you specify in your configuration file, **astilectron-bundler** will create a folder `<output path you specified in the configuration file>/<os>-<arch>` that will contain the proper files.

# Ldflags

**astilectron-bundler** uses `ldflags` when building the project. It means if you add one of the following variables as global exported variables in your project, they will have the following value:

- `AppName`:  filled with the configuration app name
- `BuiltAt`: filled with the date the build has been done at

# Subcommands
## Only bind data: bd

Use this subcommand if you want to skip most of the bundling process and only bind data/generate the `bind.go` file (useful when you want to test your app running `go run *.go`):

    $ astilectron-bundler bd -v -c <path to your configuration file>

## Clear the cache: cc

The **bundler** stores downloaded files in a cache to avoid downloading them over and over again. That cache may be corrupted. In that case, use this subcommand to clear the cache:

    $ astilectron-bundler cc -v