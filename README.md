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
    {"arch": "amd64", "os": "windows"}
  ],
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