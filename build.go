package main

import (
	"archive/zip"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// Artifact defines a build target for Go binaries.
type Artifact struct {
	Name       string `json:"name"`
	OS         string `json:"os"`
	ARCH       string `json:"arch"`
	CGOEnabled bool   `json:"cgoEnabled,omitzero"`

	Flags []string `json:"flags,omitzero"`
}

// Build calls `go build` on the artifact in the srcDir and writes the output to outDir.
//
//	ex: err := bin.Build("src", "build")
func (a *Artifact) Build(srcDir, outDir string) error {

	target := path.Join(outDir, a.Name)

	// setup the base build flags of output and target
	flags := []string{"build", "-o", target}

	// append any additional flags to the build command
	flags = append(flags, a.Flags...)

	// create the build command unrolling our flags
	cmd := exec.Command("go", flags...)
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(),
		"GOOS="+a.OS,
		"GOARCH="+a.ARCH,
	)

	// if cgo is enabled, set the env var
	if a.CGOEnabled {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
		slog.Info("executing", "GOOS", a.OS, "GOARCH", a.ARCH, "CGO_ENABLED", a.CGOEnabled, "cmd", cmd.String())
	} else {
		slog.Info("executing", "GOOS", a.OS, "GOARCH", a.ARCH, "cmd", cmd.String())
	}

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// CreateSumFile creates a checksum file for the given artifact.
//
//	ex: err := bin.CreateSumFile(sha256.New(), artifact, "example.sha256.txt")
func (b *Artifact) CreateSumFile(h hash.Hash, artifact, filename string) error {

	// open the artifact binary
	file, err := os.Open(artifact)
	if err != nil {
		return fmt.Errorf("error opening artifact %s: %w", artifact, err)
	}
	defer file.Close()

	// populate the hash with the file contents
	if _, err := io.Copy(h, file); err != nil {
		return fmt.Errorf("error hashing file %s: %w", artifact, err)
	}

	// flush out a sum of the file
	sum := h.Sum(nil)

	// follow the format the gnu core text/utilities use; eg: 'abc123 myfile'
	record := fmt.Sprintf("%x %s\n", sum, b.Name) // we use the name, not the location

	// write the line record to our sum file
	err = os.WriteFile(filename, []byte(record), 0644)
	if err != nil {
		return fmt.Errorf("error writing sum file %s: %w", filename, err)
	}

	return nil
}

// CreatZipFile creates a zip archive as filename with the contents of artifact.
//
// The artifact that is added to the zip will be at the root
// so that it can be unzipped and run from the same directory.
//
//	ex: err := bin.CreatZipFile("build/example.exe", "example.zip")
func (b *Artifact) CreatZipFile(artifact, filename string) error {

	// create the zip file
	zipFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating zipfile %s: %w", filename, err)
	}
	defer zipFile.Close()

	// create a new zip writer for the zip file
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// open the artifact file as data
	artifactData, err := os.Open(artifact)
	if err != nil {
		return fmt.Errorf("error opening artifact %s: %w", artifact, err)
	}
	defer artifactData.Close()

	// create a new zip header for the artifact
	header := &zip.FileHeader{
		Name:   filepath.Base(artifact),
		Method: zip.Deflate,
	}
	header.SetMode(0755) // make the artifact executable for unix-likes

	// create a new zip entry for the artifact using our header
	artifactEntry, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("error creating zip entry %s: %w", artifact, err)
	}

	// copy the artifact to the zip entry
	if _, err := io.Copy(artifactEntry, artifactData); err != nil {
		return fmt.Errorf("error copying artifact to zip %s: %w", artifact, err)
	}

	return nil
}
