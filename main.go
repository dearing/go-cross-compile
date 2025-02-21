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
var argVersion = flag.Bool("version", false, "emit version and build info and exit")

func usage() {
	println(`Usage: [go tool] go-cross-compile [options]
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

Options:`)
	flag.PrintDefaults()
}

func main() {
	// build systems would expect non-zero exit codes to pivot on errors
	status := run()
	os.Exit(status)
}

// run loads a config and executes the build process, returning an exit code
//
// The reason for this is to allow the main function to exit with a non-zero
// status code on error and run can return close up any scoped resources.
func run() int {

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
		config.AddBuild(name+"-darwin-amd64", "darwin", "amd64", false)
		config.AddBuild(name+"-darwin-arm64", "darwin", "arm64", false)
		config.AddBuild(name+"-linux-arm64", "linux", "arm64", false)
		config.AddBuild(name+"-linux-amd64", "linux", "amd64", false)
		config.AddBuild(name+"-linux-amd64-stripped", "linux", "amd64", false, "-ldflags=-s -w")
		config.AddBuild(name+"-windows-amd64.exe", "windows", "amd64", false)
		config.AddBuild(name+"-windows-amd64-stripped.exe", "windows", "amd64", false, "-ldflags=-s -w")
		config.AddBuild(name+"-windows-arm64.exe", "windows", "arm64", false)

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

	// check the srcDir exists
	if _, err := os.Stat(config.SrcDir); os.IsNotExist(err) {
		slog.Error("srcDir does not exist", "srcDir", config.SrcDir)
		return ErrorSrcDirNotFound
	}

	// check the outDir exists
	if _, err := os.Stat(config.OutDir); os.IsNotExist(err) {
		slog.Error("outDir does not exist", "outDir", config.OutDir)
		return ErrorOutDirNotFound
	}

	// clock the overall operation until the end
	startOperation := time.Now()

	slog.Info("building artifact", "srcDir", config.SrcDir, "outDir", config.OutDir)

	// iterate over the artifacts and call their build function
	for _, artifact := range config.Artifacts {

		// clock the build time
		start := time.Now()

		// build the artifact
		err := artifact.Build(config.SrcDir, config.OutDir)
		if err != nil {
			slog.Error("error building artifact", "error", err)
			return ErrorGoBuild
		}

		slog.Info("built", "artifact", artifact.Name, "duration", time.Since(start))

		artifactFile := fmt.Sprintf("%s/%s", config.OutDir, artifact.Name)

		// create md5 hash if requested
		if config.MD5 {
			sumFile := fmt.Sprintf("%s/%s.%s.txt", config.OutDir, artifact.Name, "md5")
			err := artifact.CreateSumFile(md5.New(), artifactFile, sumFile)
			if err != nil {
				slog.Error("error creating md5", "error", err)
				return ErrorMD5SumFile
			}

			slog.Info("created md5", "sumFile", sumFile)
		}

		// create sha1 hash if requested
		if config.SHA1 {
			sumFile := fmt.Sprintf("%s/%s.%s.txt", config.OutDir, artifact.Name, "sha1")
			err := artifact.CreateSumFile(sha1.New(), artifactFile, sumFile)
			if err != nil {
				slog.Error("error creating sha1", "error", err)
				return ErrorSHA1SumFile
			}

			slog.Info("created sha1", "sumFile", sumFile)
		}

		// create sha256 hash if requested
		if config.SHA256 {
			sumFile := fmt.Sprintf("%s/%s.%s.txt", config.OutDir, artifact.Name, "sha256")
			err := artifact.CreateSumFile(sha256.New(), artifactFile, sumFile)
			if err != nil {
				slog.Error("error creating sha256", "error", err)
				return ErrorSHA256SumFile
			}

			slog.Info("created sha256", "sumFile", sumFile)
		}

		// create sha512 hash if requested
		if config.SHA512 {
			sumFile := fmt.Sprintf("%s/%s.%s.txt", config.OutDir, artifact.Name, "sha512")
			err := artifact.CreateSumFile(sha512.New(), artifactFile, sumFile)
			if err != nil {
				slog.Error("error creating sha512", "error", err)
				return ErrorSHA512SumFile
			}

			slog.Info("created sha512", "sumFile", sumFile)
		}

		// create a zip archive of the artifact if requested
		if config.ZipFile {
			zipFile := fmt.Sprintf("%s/%s.zip", config.OutDir, artifact.Name)
			err := artifact.CreatZipFile(artifactFile, zipFile)
			if err != nil {
				slog.Error("error creating zip archive", "error", err)
				return ErrorZipFile
			}

			slog.Info("created archive", "zipFile", zipFile)
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
