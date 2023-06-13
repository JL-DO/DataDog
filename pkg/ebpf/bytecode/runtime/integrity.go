// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build ignore

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

// This program is intended to be called from go generate.
// It generates a Asset object that can be used to verify the contents
// of a file match to when go:generate was called.
func main() {
	if len(os.Args[1:]) < 3 {
		panic("please use 'go run integrity.go <in_file> <out_file> <package>'")
	}

	// cwd is guaranteed to be the directory where the go:generate comment is found
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to get current working directory: %s", err)
	}
	root := rootDir(cwd)
	args := os.Args[1:]
	inputFile, err := resolvePath(root, args[0])
	if err != nil {
		log.Fatalf("unable to resolve path to %s: %s", args[0], err)
	}
	outputFile, err := resolvePath(root, args[1])
	if err != nil {
		log.Fatalf("unable to resolve path to %s: %s", args[1], err)
	}

	err = genIntegrity(inputFile, outputFile, args[2])
	if err != nil {
		log.Fatalf("error generating integrity: %s", err)
	}
	fmt.Printf("successfully generated from %s => %s\n", inputFile, outputFile)
}

func genIntegrity(inputFile, outputFile, pkg string) error {
	hash, err := hashFile(inputFile)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return err
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	base := filepath.Base(inputFile)
	name := sanitizeFilename(strings.Title(strings.TrimSuffix(base, filepath.Ext(base))))

	imports := ""
	packagePrefix := ""

	_, curFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to get current file path")
	}

	runtimeDir := filepath.Dir(curFile)

	if filepath.Dir(outputFile) != runtimeDir {
		packagePrefix = "runtime."
		imports = "import \"github.com/DataDog/datadog-agent/pkg/ebpf/bytecode/runtime\"\n"
	}

	if err := assetTemplate.Execute(f, struct {
		Package       string
		AssetName     string
		Filename      string
		Hash          string
		Imports       string
		PackagePrefix string
	}{pkg, name, base, hash, imports, packagePrefix}); err != nil {
		return err
	}

	return nil
}

func sanitizeFilename(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '-', '_', '.':
			return -1
		default:
			return r
		}
	}, s)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("unable to read input file: %w", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("error hashing input file: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

var assetTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
//go:build linux && ebpf

package {{ .Package }}

{{ .Imports -}}

var {{ .AssetName }} = {{ .PackagePrefix }}newAsset("{{ .Filename }}", "{{ .Hash }}")
`))

// rootDir returns the base repository directory, just before `pkg`.
// If `pkg` is not found, the dir provided is returned.
func rootDir(dir string) string {
	pkgIndex := -1
	parts := strings.Split(dir, string(filepath.Separator))
	for i, d := range parts {
		if d == "pkg" {
			pkgIndex = i
			break
		}
	}
	if pkgIndex == -1 {
		return dir
	}
	return strings.Join(parts[:pkgIndex], string(filepath.Separator))
}

func resolvePath(root, path string) (string, error) {
	if strings.HasPrefix(path, "pkg/") {
		return filepath.Join(root, path), nil
	}
	return filepath.Abs(path)
}
