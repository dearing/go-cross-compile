package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

var argConfigFile = flag.String("config-file", "go-cross-compile.json", "config file to use")
var argInitConfig = flag.Bool("init-config", false, "initialize a new config file and exit")
var argSkipBuild = flag.Bool("skip-build", false, "skip the build step")
var argVersion = flag.Bool("version", false, "emit version and build info and exit")

func usage() {
	println(`Usage: go-cross-compile [options]
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
  - the zipFile will contain the binary at the root of the tree
  - the hash sum text file are compatible with the gnu hash utility suite
  - the argument --version will emit debug metadata of the tool itself

Options:`)
	flag.PrintDefaults()
}

func main() {
	// build systems would expect non-zero exit codes to pivot on errors
	status := work()
	os.Exit(status)
}

func work() int {

	flag.Usage = usage
	flag.Parse()

	if *argVersion {
		VersionInfo()
		return NoError
	}

	config := NewConfig()

	// generate a starter config and exit
	if *argInitConfig {

		name := "example"

		// try to get the current working directory as a default name
		pwd, err := os.Getwd()
		if err == nil {
			name = filepath.Base(pwd)
		}

		// add some common build targets
		config.AddBuild(name+"-darwin-amd64", "darwin", "amd64")
		config.AddBuild(name+"-darwin-arm64", "darwin", "arm64")
		config.AddBuild(name+"-linux-amd64", "linux", "amd64")
		config.AddBuild(name+"-linux-arm64", "linux", "arm64")
		config.AddBuild(name+"-windows-amd64.exe", "windows", "amd64")
		config.AddBuild(name+"-windows-arm64.exe", "windows", "arm64")

		// clobber a config file for the user with defaults
		err = config.Save(*argConfigFile)
		if err != nil {
			slog.Error("error saving config", "config-file", *argConfigFile, "error", err)
			return ErrorInitConfig
		}
		slog.Info("created new config", "config-file", *argConfigFile)

		return NoError
	}

	// load the config file
	if err := config.Load(*argConfigFile); err != nil {
		slog.Error("error loading config", "config-file", *argConfigFile, "error", err)
		return ErrorConfigFileNotFound
	}

	// run some basic checks on the config
	config.RunChecks()

	// check the src-dir exists
	if _, err := os.Stat(config.SrcDir); os.IsNotExist(err) {
		slog.Error("src-dir does not exist", "src-dir", config.SrcDir)
		return ErrorSrcDirNotFound
	}

	// check the outDir exists
	if _, err := os.Stat(config.OutDir); os.IsNotExist(err) {
		slog.Error("outDir does not exist", "outDir", config.OutDir)
		return ErrorOutDirNotFound
	}

	// clock the overall operation until the end
	startOperation := time.Now()

	// TODO: this still feels clunky
	if *argSkipBuild {

		slog.Info("skipping build step")

	} else {

		slog.Info("building binaries", "src-dir", config.SrcDir, "outDir", config.OutDir)

		// iterate over the binaries and call their build function
		for _, binary := range config.Artifacts {

			// clock the build time
			start := time.Now()

			// build the binary
			err := binary.Build(config.SrcDir, config.OutDir)
			if err != nil {
				slog.Error("error building binary", "error", err)
				return ErrorGoBuild
			}

			slog.Info("built", "binary", binary.Name, "duration", time.Since(start))

			artifact := fmt.Sprintf("%s/%s", config.OutDir, binary.Name)

			// md5 if requested
			if config.MD5 {
				sumFile := fmt.Sprintf("%s/%s.%s.txt", config.OutDir, binary.Name, "md5")
				err := binary.CreateSumFile(md5.New(), artifact, sumFile)
				if err != nil {
					slog.Error("error creating md5", "error", err)
					return ErrorMD5SumFile
				}

				slog.Info("created md5", "sum-file", sumFile)
			}

			// sha1 if requested
			if config.SHA1 {
				sumFile := fmt.Sprintf("%s/%s.%s.txt", config.OutDir, binary.Name, "sha1")
				err := binary.CreateSumFile(sha1.New(), artifact, sumFile)
				if err != nil {
					slog.Error("error creating sha1", "error", err)
					return ErrorSHA1SumFile
				}

				slog.Info("created sha1", "sum-file", sumFile)
			}

			// sha256 if requested
			if config.SHA256 {
				sumFile := fmt.Sprintf("%s/%s.%s.txt", config.OutDir, binary.Name, "sha256")
				err := binary.CreateSumFile(sha256.New(), artifact, sumFile)
				if err != nil {
					slog.Error("error creating sha256", "error", err)
					return ErrorSHA256SumFile
				}

				slog.Info("created sha256", "sum-file", sumFile)
			}

			// sha512 if requested
			if config.SHA512 {
				sumFile := fmt.Sprintf("%s/%s.%s.txt", config.OutDir, binary.Name, "sha512")
				err := binary.CreateSumFile(sha512.New(), artifact, sumFile)
				if err != nil {
					slog.Error("error creating sha512", "error", err)
					return ErrorSHA512SumFile
				}

				slog.Info("created sha512", "sum-file", sumFile)
			}

			// zip if requested
			if config.ZipFile {
				zipFile := fmt.Sprintf("%s/%s.zip", config.OutDir, binary.Name)
				err := binary.CreatZipFile(artifact, zipFile)
				if err != nil {
					slog.Error("error creating zip", "error", err)
					return ErrorZipFile
				}

				slog.Info("created zip", "zip-file", zipFile)
			}
		}
	}

	slog.Info("operation complete", "duration", time.Since(startOperation))
	return NoError
}

func VersionInfo() {
	// seems like a nice place to sneak in some debug information
	info, ok := debug.ReadBuildInfo()
	if ok {
		slog.Info("build info", "main", info.Main.Path, "version", info.Main.Version)
		for _, setting := range info.Settings {
			slog.Info("build info", "key", setting.Key, "value", setting.Value)
		}
	}
}
