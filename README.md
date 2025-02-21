# go-cross-compile

simple go tool to cross-compile binaries, hash sums and archives for publishing

This tool doesn't rely anything outside of the standard library, the hashing and compression are handled internally. This keeps the tool nimble and cross platform meaning we don't need to rely on build environments prepopulated with packages (aside from Go itself).

## global install

```
go install github.com/dearing/go-cross-compile
```
>[!NOTE]
>With Go 1.24+, we can use the tool feature to pin the tool in go.mod instead of installing it on the host's path
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
## 30 second demo
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
Usage: go-cross-compile [options]
This tool was inspired from the tedious task of cross compiling go binaries and
then uploading them to github. The md5, sha1, sha256, sha512, and zip options
are available for each artifact of the build and you end up with the following:

  $outDir/$name per artifact member of artifacts in the config
  $outDir/$name$hash.txt for each of md5, sha1, sha256 or sha512 when enabled
  $outDir/$name.zip if zipFile when enabled

Workflow:

  1. generate a new config file with 'go-cross-compile --init-config'
  2. edit the config file 'go-cross-compile.json' to your liking
  3. create the outDir ex: 'mkdir build' (this is a kind of safety check)
  4. run 'go-cross-compile --config-file go-cross-compile.json' to do work

Tips:
  - 'go tool dist list' will show the valid GOOS and GOARCH values
  - zipFile will contain on the binary at the root of the tree
  - the hash files are in the format of the *nix hash utilities
  - the argument --version will emit debug metadata of the tool itself

Options:
  -config-file string
        config file to use (default "go-cross-compile.json")
  -init-config
        initialize a new config file and exit
  -skip-build
        skip the build step
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
      "name": "go-cross-compile-linux-amd64",
      "os": "linux",
      "arch": "amd64"
    },
    {
      "name": "go-cross-compile-linux-arm64",
      "os": "linux",
      "arch": "arm64"
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
