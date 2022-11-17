package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/kardianos/service"

	pkglog "servicemanager/pkg/log"
	"servicemanager/pkg/manager"
	"servicemanager/pkg/util"
)

var action string

func init() {
	flag.StringVar(&action, "action", "", "start|stop|restart|install|uninstall")
}

func main() {
	go func() {
		http.ListenAndServe("localhost:9001", nil)
	}()

	flag.Parse()

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
