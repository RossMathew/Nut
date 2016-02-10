package specification

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Group struct {
	Members map[string]Member
}

func GroupFromYAML(file string) (*Group, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var g Group
	if err := yaml.Unmarshal(data, &g.Members); err != nil {
		return nil, err
	}
	return &g, nil
}

func (g *Group) Create() error {
	for name, member := range g.Members {
		if err := member.Create(name); err != nil {
			return err
		}
		if err := member.RunCommand(); err != nil {
			return err
		}
	}
	return nil
}
