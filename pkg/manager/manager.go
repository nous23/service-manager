package manager

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"

	"servicemanager/pkg/util"
)

const (
	Name        = "ServiceManager"
	DisplayName = "Service Manager"
	Description = `Service Manager manages user custom services, providing a web UI to control services.`
)

func New() (service.Service, error) {
	config := &service.Config{
		Name:        Name,
		DisplayName: DisplayName,
		Description: Description,
		Arguments: []string{
			"--log-level=debug",
		},
	}

	sm := &serviceManager{
		configReload: make(chan struct{}, 100),
		tasks:        make(map[string]*Task),
	}

	r, err := newServer(sm)
	if err != nil {
		log.Errorf("create server failed: %v", err)
		return nil, err
	}
	sm.server = r
	s, err := service.New(sm, config)
	if err != nil {
		log.Errorf("create service failed: %v", err)
		return nil, err
	}
	return s, nil
}

type serviceManager struct {
	configWatcher *fsnotify.Watcher
	tasks         map[string]*Task
	configReload  chan struct{}
	config        *Configs
	server        *gin.Engine
}

type Task struct {
	Name    string
	binPath string
	args    []string
	cmd     *exec.Cmd
	Config  *TaskConfig
}

type TaskList []*Task

func (t *Task) String() string {
	return fmt.Sprintf("[%s] %s %s", t.Name, t.binPath, strings.Join(t.args, " "))
}

func (t *Task) start() {
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

func (t *Task) stop() {
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
		log.Infof("watch config write")
		sm.configReload <- struct{}{}
	default:
		log.Infof("watch config %s", event.String())
	}
}

func (sm *serviceManager) run() {
	log.Debug("run service manager")
	go func() {
		err := sm.server.Run(":9999")
		if err != nil {
			log.Errorf("run server failed: %v", err)
		}
		log.Errorf("server exit")
	}()

	configs, err := loadConfig()
	if err != nil {
		log.Errorf("load config failed: %v", err)
	}
	sm.config = configs
	tm := make(map[string]*Task)
	for _, taskConf := range sm.config.Tasks {
		t, err := createTask(taskConf)
		if err != nil {
			log.Errorf("create task failed: %v", err)
			continue
		}
		if tm[t.Name] != nil {
			log.Errorf("task %s already exits: %s", t.Name, tm[t.Name].String())
			continue
		}
		tm[t.Name] = t
	}
	sm.tasks = tm
	for _, t := range sm.tasks {
		t.start()
	}

	for {
		select {
		case <-sm.configReload:
			configs, err = loadConfig()
			if err != nil {
				log.Warningf("load config failed: %v", err)
				continue
			}

			newTasks := make(map[string]*TaskConfig)
			for _, tc := range configs.Tasks {
				newTasks[tc.Name] = tc
			}

			// stop deleted tasks
			var toStop []string
			for _, t := range sm.tasks {
				_, ok := newTasks[t.Name]
				if !ok {
					toStop = append(toStop, t.Name)
				}
			}
			for _, taskName := range toStop {
				sm.tasks[taskName].stop()
				delete(sm.tasks, taskName)
			}

			// start new task and restart updated task
			for _, tc := range configs.Tasks {
				oldTask, ok := sm.tasks[tc.Name]
				if ok {
					if oldTask.Config.Equivalent(tc) {
						log.Infof("task %s not changed, keep running", tc.Name)
						continue
					}
					// update task
					oldTask.stop()
					newTask, err := createTask(tc)
					if err != nil {
						log.Errorf("create task failed: %v", err)
						continue
					}
					newTask.start()
					sm.tasks[tc.Name] = newTask
					continue
				} else {
					// start new task
					newTask, err := createTask(tc)
					if err != nil {
						log.Errorf("create task failed: %v", err)
						continue
					}
					newTask.start()
					sm.tasks[tc.Name] = newTask
				}
			}
		}
	}
}

func (sm *serviceManager) getTaskList() TaskList {
	var ts = TaskList{}
	for _, t := range sm.tasks {
		ts = append(ts, t)
	}
	return ts
}

func createTask(c *TaskConfig) (*Task, error) {
	if c.Name == "" {
		return nil, fmt.Errorf("task name not specified")
	}
	if c.BinPath == "" {
		return nil, fmt.Errorf("bin path not specified")
	}
	return &Task{
		Name:    c.Name,
		binPath: c.BinPath,
		args:    c.Args,
		Config:  c,
	}, nil
}
