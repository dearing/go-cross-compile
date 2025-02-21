# go-cross-compile

simple go tool to cross-compile binaries, hash sums and archives for publishing

This tool doesn't rely anything outside of the standard library, the hashing and compression are handled internally. This keeps the tool nimble and cross platform meaning we don't need to rely on build environments prepopulated with packages (aside from Go itself). You can customize each build with additional flags and CGO_ENABLED environment variable as needed, see the example config below.

## global install

```
go install github.com/dearing/go-cross-compile
```
>[!NOTE]
>With Go 1.24+, we can use the tool feature to pin the tool in go.mod instead of installing it in the host's path
## go tool usage
```
go get -tool github.com/dearing/go-cross-compile@latest
go tool go-cross-compile --version
go tool go-cross-compile --init-config
go tool go-cross-compile
```
## tool maintenance tips
```
      pin => go get -tool github.com/dearing/go-cross-compile@v1.0.1
   update => go get -tool github.com/dearing/go-cross-compile@latest
downgrade => go get -tool github.com/dearing/go-cross-compile@v1.0.0
uninstall => go get -tool github.com/dearing/go-cross-compile@none
```
## 30 second test drive
```
git clone https://github.com/dearing/go-cross-compile.git
cd go-cross-compile
mkdir build
go tool go-cross-compile --config-file .github/go-cross-compile.json
dir build
```
---
## usage

```
Usage: [go tool] go-cross-compile [options]
This tool was inspired from the tedious task of cross compiling go binaries and
hash sums. The md5, sha1, sha256, sha512 and zip options are available for each
artifact of the operation.

  - $outDir/$name per artifact member of artifacts defined in the config
  - $outDir/$name.$mode.txt for each of md5, sha1, sha256 or sha512 when enabled
  - $outDir/$name.zip containing the artifact per build when enabled

Workflow:

  1. generate a new config file with 'go-cross-compile --init-config'
  2. edit the config file 'go-cross-compile.json' to your liking
  3. create the outDir ex: 'mkdir build' (this is a kind of safety check)
  4. run 'go-cross-compile --config-file go-cross-compile.json' to have Go build
         the artifacts and create the hash sums and zip files

Tips:
  - 'go tool dist list' will show the valid GOOS and GOARCH values
  - the zip files will contain the artifact at the root of the tree
  - the hash sum text files are compatible with the gnu core text utilities
  - the argument --version will emit debug metadata of the tool itself
  - the tool will exit with a non-zero status on error

Options:
  -config-file string
        config file to use (default "go-cross-compile.json")
  -init-config
        initialize a new config file and exit
  -version
        emit version and build info and exit
```
## example config
```json
{
  "outDir": "build",
  "srcDir": ".",
  "md5": false,
  "sha1": true,
  "sha256": false,
  "sha512": false,
  "zipFile": false,
  "artifacts": [
    {
      "name": "go-cross-compile-darwin-amd64",
      "os": "darwin",
      "arch": "amd64"
    },
    {
      "name": "go-cross-compile-darwin-arm64",
      "os": "darwin",
      "arch": "arm64"
    },
    {
      "name": "go-cross-compile-linux-arm64",
      "os": "linux",
      "arch": "arm64"
    },
    {
      "name": "go-cross-compile-linux-amd64",
      "os": "linux",
      "arch": "amd64",
      "cgoEnabled": true,
      "flags": [
        "-race"
      ]
    },
    {
      "name": "go-cross-compile-windows-amd64.exe",
      "os": "windows",
      "arch": "amd64"
    },
    {
      "name": "go-cross-compile-windows-arm64.exe",
      "os": "windows",
      "arch": "arm64"
    }
  ]
}
```
