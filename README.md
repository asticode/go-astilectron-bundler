This package provides a way to bundle Astilectron apps properly.

It has only been tested in MacOS for now, and is still a work in progress.

# Installation

Run the following command:

    $ go get -u github.com/asticode/go-astilectron-bundler/...
    
# Configuration

**astilectron-bundler** uses a configuration file to know what it's supposed to do. Here's an example:

```json
{
  "app_name": "Test",
  "app_icon_darwin_path": "path/to/icon.icns",
  "app_icon_default_path": "path/to/icon.png",
  "environments": [
    {"arch": "amd64", "os": "darwin"},
    {"arch": "amd64", "os": "linux"},
    {"arch": "amd64", "os": "windows"}
  ],
  "input_path": "path/to/src/github.com/username/project",
  "output_path": "path/to/output/directory"
}
```

Paths can be either relative or absolute but we **strongly** encourage to use relative paths.

If no input path is specified, the working directory path is used. We **strongly** encourage to `d` into your input path before executing the bundler.

# Usage

If **astilectron-bundler** has been installed properly (and the GOPATH is in your PATH), run the following command:

    $ astilectron-bundler -v -c <path to your configuration file>
    
# Output

For each environment you specified in your configuration file, **astilectron-bundler** will create a folder at `<output path you specified in the configuration file>/<os>/<arch>`.

Depending on the OS and the arch you specified, you'll find the proper files in here.

# CLI flags

The available CLI flags are:

- `a`: the bundled environment is the current one
- `v`: debug logs are shown

# ldflags

**astilectron-bundler** uses `ldflags` when building the project. The following variables are set:

- `AppName`:  filled with the configuration app name
- `BuiltAt`: filled with the date the build has been done at

and can be used in your project.