package main

import (
	"github.com/alecthomas/kingpin"

	pkglog "servicemanager/pkg/log"
	"servicemanager/pkg/manager"
	"servicemanager/pkg/util"
)

func main() {
	kingpin.Parse()

	err := pkglog.Init()
	if err != nil {
		util.ExitOnError(err)
	}

	m, err := manager.New()
	if err != nil {
		util.ExitOnError(err)
	}
	err = m.Run()
	if err != nil {
		util.ExitOnError(err)
	}
}
