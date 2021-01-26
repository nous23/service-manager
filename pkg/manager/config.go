package manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

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
		currDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Errorf("get current directory failed: %v", err)
			return err
		}
		configPath = filepath.Join(currDir, configFileName)
	}
	if !util.Exists(configPath) {
		return fmt.Errorf("config file %s does not exists", configPath)
	}
	return nil
}

func loadConfig() (*Configs, error) {
	if configPath == "" {
		currDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Errorf("get current directory failed: %v", err)
			return nil, err
		}
		configPath = filepath.Join(currDir, configFileName)
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
