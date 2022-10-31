// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux_bpf
// +build linux_bpf

package ebpf

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf/btf"
	"github.com/mholt/archiver/v3"

	"github.com/DataDog/datadog-agent/pkg/metadata/host"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

func GetBTF(userProvidedBtfPath, bpfDir string) (*btf.Spec, COREResult) {
	var btfSpec *btf.Spec
	var err error

	if userProvidedBtfPath != "" {
		btfSpec, err = loadBTFFrom(userProvidedBtfPath)
		if err == nil {
			log.Debugf("loaded BTF from %s", userProvidedBtfPath)
			return btfSpec, successCustomBTF
		}
	}

	btfSpec, err = checkEmbeddedCollection(filepath.Join(bpfDir, "co-re/btf/"))
	if err == nil {
		log.Debugf("loaded BTF from embedded collection")
		return btfSpec, successEmbeddedBTF
	}
	log.Debugf("couldn't find BTF in embedded collection: %s", err)

	btfSpec, err = btf.LoadKernelSpec()
	if err == nil {
		log.Debugf("loaded BTF from default kernel location")
		return btfSpec, successDefaultBTF
	}
	log.Debugf("couldn't find BTF in default kernel locations: %s", err)

	return nil, btfNotFound
}

func checkEmbeddedCollection(collectionPath string) (*btf.Spec, error) {
	si := host.GetStatusInformation()
	platform := si.Platform
	platformVersion := si.PlatformVersion
	kernelVersion := si.KernelVersion

	btfSubdirectory := filepath.Join(platform)
	if platform == "ubuntu" {
		// Ubuntu BTFs are stored in subdirectories corresponding to platform version.
		// This is because we have BTFs for different versions of ubuntu with the exact same
		// kernel name, so kernel name alone is not a unique identifier.
		btfSubdirectory = filepath.Join(platform, platformVersion)
	}

	// If we've previously extracted the BTF file in question, we can just load it
	extractedBtfPath := filepath.Join(collectionPath, btfSubdirectory, kernelVersion+".btf")
	if _, err := os.Stat(extractedBtfPath); err == nil {
		return loadBTFFrom(extractedBtfPath)
	}
	log.Debugf("extracted btf file not found at %s: attempting to extract from embedded archive", extractedBtfPath)

	// The embedded BTFs are compressed twice: the individual BTFs themselves are compressed, and the collection
	// of BTFs as a whole is also compressed.
	// This means that we'll need to first extract the specific BTF which  we're looking for from the collection
	// tarball, and then unarchive it.
	btfTarball := filepath.Join(collectionPath, btfSubdirectory, kernelVersion+".btf.tar.xz")
	if _, err := os.Stat(btfTarball); errors.Is(err, fs.ErrNotExist) {
		collectionTarball := filepath.Join(collectionPath, "minimized-btfs.tar.xz")
		targetBtfFile := filepath.Join(btfSubdirectory, kernelVersion+".btf.tar.xz")

		if err := archiver.NewTarXz().Extract(collectionTarball, targetBtfFile, collectionPath); err != nil {
			return nil, err
		}
	}

	destinationFolder := filepath.Join(collectionPath, btfSubdirectory)
	if err := archiver.NewTarXz().Unarchive(btfTarball, destinationFolder); err != nil {
		return nil, err
	}
	return loadBTFFrom(filepath.Join(destinationFolder, kernelVersion+".btf"))
}

func loadBTFFrom(path string) (*btf.Spec, error) {
	data, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return btf.LoadSpecFromReader(data)
}
