package manager

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"

	"servicemanager/pkg/util"
)

func New() (service.Service, error) {
	u, err := user.Current()
	if err != nil {
		log.Errorf("get current user failed: %v", err)
		return nil, err
	}
	config := &service.Config{
		Name:        "ServiceManager",
		DisplayName: "Service Manager",
		Description: "Manage Windows Service",
		UserName:    u.Username,
		Option:      nil,
	}

	sm := &serviceManager{
		configReload: make(chan struct{}, 100),
		tasks:        make(map[string]*task),
	}
	s, err := service.New(sm, config)
	if err != nil {
		log.Errorf("create service failed: %v", err)
		return nil, err
	}
	return s, nil
}

type serviceManager struct {
	configWatcher *fsnotify.Watcher
	tasks         map[string]*task
	configReload  chan struct{}
	config        *Configs
}

type task struct {
	name    string
	binPath string
	args    []string
	cmd     *exec.Cmd
}

func (t *task) String() string {
	return fmt.Sprintf("[%s] %s %s", t.name, t.binPath, strings.Join(t.args, " "))
}

func (t *task) start() {
	log.Debugf("start task %s", t.String())
	c := exec.Command(t.binPath, t.args...)
	go func() {
		output, err := c.CombinedOutput()
		if err != nil {
			log.Errorf("run command %s failed: %s", c.String(), string(output))
		}
	}()
	t.cmd = c
}

func (t *task) stop() {
	log.Debugf("stop task %s", t.String())
	if t.cmd == nil {
		return
	}
	err := t.cmd.Process.Kill()
	if err == nil {
		log.Infof("stop command %s success", t.cmd.String())
		return
	}
	log.Warningf("kill process %d failed: %v", t.cmd.Process.Pid, err)
	// try force kill process
	cmd := exec.Command("taskkill", "/f", "/pid", fmt.Sprintf("%d", t.cmd.Process.Pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("stop command %s failed: %v, output: %s", t.cmd.String(), err, string(output))
	}
}

func (sm *serviceManager) Start(service.Service) error {
	go sm.start()
	return nil
}

func (sm *serviceManager) Stop(service.Service) error {
	sm.stop()
	return nil
}

func (sm *serviceManager) start() {
	log.Debug("start service manager")
	err := initConfigPath()
	if err != nil {
		util.ExitOnError(err)
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		util.ExitOnError(err)
	}
	err = w.Add(configPath)
	if err != nil {
		util.ExitOnError(err)
	}
	sm.configWatcher = w
	go sm.watchConfig()
	sm.run()
	neverExit := make(chan int)
	<-neverExit
}

func (sm *serviceManager) stop() {
	for _, t := range sm.tasks {
		t.stop()
	}
}

func (sm *serviceManager) watchConfig() {
	for {
		select {
		case event, ok := <-sm.configWatcher.Events:
			if !ok {
				return
			}
			log.Infof("event: %v", event)
			sm.dealWithConfigOperation(&event)
		case err, ok := <-sm.configWatcher.Errors:
			if !ok {
				return
			}
			log.Errorf("error: %v", err)
		}
	}
}

func (sm *serviceManager) dealWithConfigOperation(event *fsnotify.Event) {
	if event == nil {
		return
	}
	switch event.Op {
	case fsnotify.Write:
		sm.configReload <- struct{}{}
	default:

	}
}

func (sm *serviceManager) run() {
	log.Debug("run service manager")
	configs, err := loadConfig()
	if err != nil {
		log.Errorf("load config failed: %v", err)
	}
	sm.config = configs
	tm := make(map[string]*task)
	for _, taskConf := range sm.config.Tasks {
		t := &task{
			name:    taskConf.Name,
			binPath: taskConf.BinPath,
			args:    taskConf.Args,
		}
		if tm[t.name] != nil {
			log.Errorf("task %s already exits: %s", t.name, tm[t.name].String())
			continue
		}
		tm[t.name] = t
	}
	sm.tasks = tm
	for _, t := range sm.tasks {
		t.start()
	}
}
