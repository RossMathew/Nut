package specification

import (
	"path/filepath"
)

type Member struct {
	Build       string
	Command     string
	Volumes     []string
	User        string
	Environment []string
	Ports       []string
}

func (m *Member) RunCommand() error {
	return nil
}
func (m *Member) Create(name string) error {
	if m.Build == "" {
		return nil
	}
	file, err := expandPath(m.Build)
	if err != nil {
		return err
	}
	spec := New(name)
	if err := spec.Parse(file); err != nil {
		return err
	}
	if err := spec.Build(m.Volumes...); err != nil {
		return err
	}
	return nil
}

func expandPath(file string) (string, error) {
	f, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return filepath.Join(f, "Dockerfile"), nil
}
