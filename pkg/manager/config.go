package manager

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"servicemanager/pkg/global"
	"servicemanager/pkg/util"
)

const configFileName = "tasks.yaml"

var configPath string

func init() {
	kingpin.Flag("config-path", "path of task config file").Default("").StringVar(&configPath)
}

// Config of task
type TaskConfig struct {
	Name    string   `yaml:"name"`
	BinPath string   `yaml:"binPath"`
	Args    []string `yaml:"args"`
}

type Configs struct {
	Tasks []*TaskConfig `yaml:"tasks"`
}

func (t *TaskConfig) Equivalent(tt *TaskConfig) bool {
	if tt == nil {
		return false
	}
	if t.BinPath != tt.BinPath {
		return false
	}
	if strings.Join(t.Args, "") != strings.Join(tt.Args, "") {
		return false
	}
	return true
}

func initConfigPath() error {
	if configPath == "" {
		configPath = filepath.Join(global.CurrDir, configFileName)
	}
	if !util.Exists(configPath) {
		return fmt.Errorf("config file %s does not exists", configPath)
	}
	return nil
}

func loadConfig() (*Configs, error) {
	if configPath == "" {
		configPath = filepath.Join(global.CurrDir, configFileName)
	}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("read config file %s failed: %v", configPath, err)
		return nil, err
	}
	configs := &Configs{}
	err = yaml.Unmarshal(data, configs)
	if err != nil {
		log.Errorf("unmarshal config file data failed: %v", err)
		return nil, err
	}
	return configs, nil
}
