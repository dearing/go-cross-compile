package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
)

// Config defines a set of build targets and options.
type Config struct {
	OutDir string `json:"outDir"`
	SrcDir string `json:"srcDir"`

	MD5     bool `json:"md5"`
	SHA1    bool `json:"sha1"`
	SHA256  bool `json:"sha256"`
	SHA512  bool `json:"sha512"`
	ZipFile bool `json:"zipFile"`

	Artifacts []Artifact `json:"artifacts"`
}

// NewConfig returns a new Config with default values
func NewConfig() *Config {
	return &Config{
		OutDir:    "build",
		SrcDir:    ".",
		MD5:       false,
		SHA1:      true,
		SHA256:    false,
		SHA512:    false,
		ZipFile:   false,
		Artifacts: []Artifact{},
	}
}

// AddBuild adds a build target group for Go binaries.
//
//	ex: myconfig.AddBuild("example.exe", "windows", "amd64")
func (c *Config) AddBuild(name, os, arch string, cgo bool, flags ...string) {
	c.Artifacts = append(c.Artifacts,
		Artifact{
			Name:       name,
			OS:         os,
			ARCH:       arch,
			CGOEnabled: cgo,
			Flags:      flags,
		})
}

// Save saves a json representation of Config to filename
//
//	ex: myconfig.Save("go-github-releaser.json")
func (c *Config) Save(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	// should we assume to clobber the file?
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Load reads filename into a Config struct
//
//	ex: myconfig.Load("go-github-releaser.json")
func (c *Config) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return nil
}

// RunChecks performs some basic checks on the config to catch gotchas.
//
//	ex: myconfig.RunChecks()
func (c *Config) RunChecks() {
	// check if arch is not found in the artifact name
	for _, artifact := range c.Artifacts {
		if !strings.Contains(artifact.Name, artifact.ARCH) {
			slog.Warn("arch not found in artifact name", "artifact", artifact.Name, "arch", artifact.ARCH)
		}
	}
}
