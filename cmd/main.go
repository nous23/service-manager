package main

import (
	"fmt"

	"github.com/alecthomas/kingpin"
	"github.com/kardianos/service"

	pkglog "servicemanager/pkg/log"
	"servicemanager/pkg/manager"
	"servicemanager/pkg/util"
)

var action string

func init() {
	kingpin.Flag("action", "start|stop|restart|install|uninstall").Default("").StringVar(&action)
}

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

	if action != "" {
		err = service.Control(m, action)
		if err != nil {
			util.ExitOnError(fmt.Errorf("%s %s failed: %v", action, manager.Name, err))
		}
		util.ExitOnSuccess(fmt.Sprintf("%s %s success", action, manager.Name))
	}
	err = m.Run()
	if err != nil {
		util.ExitOnError(err)
	}
}
