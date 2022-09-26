package main

import (
	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/pkg"
)

var (
	Version           string
	Environment       string
	RemoteVersionURL  string
	LatestDownloadURL string
)

func init() {
	newPackage := pkg.PackageInfo{
		Version:           Version,
		Environment:       Environment,
		RemoteVersionURL:  RemoteVersionURL,
		LatestDownloadURL: LatestDownloadURL,
	}

	if err := newPackage.ComputeHash(); err != nil {
		log.Error(err)
	}

	pkg.SetGlobalInfo(newPackage)
}
