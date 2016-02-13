package specification

import (
	"gopkg.in/lxc/go-lxc.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type Manifest struct {
	Labels       map[string]string
	Maintainers  []string
	ExposedPorts []uint64
	EntryPoint   []string
	Env          []string
	User         string
	WorkDir      string
}

func (m *Manifest) Load(name string) error {
	lxcdir := lxc.GlobalConfigItem("lxc.lxcpath")
	manifestPath := filepath.Join(lxcdir, name, "manifest.yml")
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &m)
}
