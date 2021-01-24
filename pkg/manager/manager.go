package manager

import (
	"os/user"

	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
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

	sm := &serviceManager{}
	s, err := service.New(sm, config)
	if err != nil {
		log.Errorf("create service failed: %v", err)
		return nil, err
	}
	return s, nil
}

type serviceManager struct {
}

func (sm *serviceManager) Start(s service.Service) error {
	// todo
	return nil
}

func (sm *serviceManager) Stop(s service.Service) error {
	// todo
	return nil
}
