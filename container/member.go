package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

// Member represents an individual member in a container group
type Member struct {
	Build         string
	ContainerName string `yaml:"container_name"`
	Hostname      string `yaml:"hostname"`
	Command       string
	Volumes       []string
	User          string
	Environment   []string
	Ports         []string
	Image         string
	ct            *Container
}

// Create creates a container from member specification
func (m *Member) Create(name string) error {
	b := NewBuilder(name)
	b.Volumes = m.Volumes
	if m.Image != "" {
		log.Debugln("Creating container from image", m.Image)
		c, err := b.CreateContainer(m.Image)
		if err != nil {
			return err
		}
		if err := c.WriteManifest(); err != nil {
			return err
		}
		m.ct = c
	}
	if m.Build == "" {
		return nil
	}
	file, err := expandPath(m.Build)
	if err != nil {
		return err
	}
	if err := b.Parse(file); err != nil {
		return err
	}
	c, err := b.Build()
	if err != nil {
		return err
	}
	m.ct = c
	return nil
}

func expandPath(file string) (string, error) {
	f, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return filepath.Join(f, "Dockerfile"), nil
}

// RunCommand runs the member's specified command inside it representative container
func (m *Member) RunCommand() error {
	if m.ct == nil {
		return fmt.Errorf("Container for member '%s' has not been created yet", m.ContainerName)
	}
	return m.ct.RunCommand(strings.Fields(m.Command))
}
