package service

// Config of services
type Config struct {
	Name    string   `yaml:"name"`
	BinPath string   `yaml:"binPath"`
	Args    []string `yaml:"args"`
}
