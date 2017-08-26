This package provides a way to bundle Astilectron apps properly.

It has only been tested in MacOS for now, and is still work in progress.

# Installation

Run the following command:

    $ go get -u github.com/asticode/go-astilectron-bundler/...
    
# Configuration

**astilectron-bundler** uses a configuration file to know what it's supposed to do. Here's an example:

```json
{
  "app_name": "Test",
  "app_icon_darwin_path": "/absolute/path/to/icon.icns",
  "app_icon_default_path": "/absolute/path/to/icon.png",
  "environments": [
    {"arch": "amd64", "os": "darwin"},
    {"arch": "amd64", "os": "linux"},
    {"arch": "amd64", "os": "windows"}
  ],
  "input_path": "/absolute/go/path/src/github.com/username/project",
  "output_path": "/absolute/path/to/output/directory"
}
```

# Usage

If **astilectron-bundler** has been installed properly (and the GOPATH is in your PATH), run the following command:

    $ astilectron-bundler -v -c <path to your configuration file>
    
# Output

For each environment you specified in your configuration file, **astilectron-bundler** will create a folder at `<output path you specified in the configuration file>/<os>/<arch>`.

Depending on the OS and the arch you specified, you'll find the proper files in here.

# ldflags

**astilectron-bundler** uses `ldflags` when building the project. The following variables are set:

- `AppName`:  filled with the configuration app name
- `BuiltAt`: filled with the date the build has been done at