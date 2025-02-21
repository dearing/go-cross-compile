package main

import (
	"archive/zip"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// Artifact defines a build target for Go binaries.
type Artifact struct {
	Name string `json:"name"`
	OS   string `json:"os"`
	ARCH string `json:"arch"`
}

// Build calls `go build` on the artifact in the srcDir and writes the output to outDir.
//
//	ex: err := bin.Build("src", "build")
func (a *Artifact) Build(srcDir, outDir string) error {

	target := path.Join(outDir, a.Name)

	cmd := exec.Command("go", "build", "-o", target)
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(),
		"GOOS="+a.OS,
		"GOARCH="+a.ARCH,
	)

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

	// open the binary file
	file, err := os.Open(artifact)
	if err != nil {
		return fmt.Errorf("error opening artifact %s: %w", artifact, err)
	}
	defer file.Close()

	// populate the hash with the file contents
	if _, err := io.Copy(h, file); err != nil {
		return fmt.Errorf("error hashing file %s: %w", artifact, err)
	}

	// flush out the hash
	sum := h.Sum(nil)

	// follow the format the *nix hash utilities use
	content := fmt.Sprintf("%x %s\n", sum, b.Name)

	// write the hash file
	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing sum file %s: %w", filename, err)
	}

	return nil
}

// CreatZipFile creates a zip file as the filename with the contents of artifact.
//
// The artifact that is added to the zip will be at the root
// so that it can be unzipped and run from one dir.
//
//	ex: err := bin.CreatZipFile("build/example.exe", "example.zip")
func (b *Artifact) CreatZipFile(artifact, filename string) error {

	// create the zip file
	zipFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating zipfile %s: %w", filename, err)
	}
	defer zipFile.Close()

	// create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// open the artifact file
	fileToZip, err := os.Open(artifact)
	if err != nil {
		return fmt.Errorf("error opening artifact %s: %w", artifact, err)
	}
	defer fileToZip.Close()

	// create a new zip entry as the basename of the artifact
	w, err := zipWriter.Create(filepath.Base(artifact))
	if err != nil {
		return fmt.Errorf("error creating zip entry %s: %w", artifact, err)
	}

	// copy the artifact to the zip entry
	if _, err := io.Copy(w, fileToZip); err != nil {
		return fmt.Errorf("error copying artifact to zip %s: %w", artifact, err)
	}

	return nil
}
