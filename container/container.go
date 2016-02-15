package container

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/lxc/go-lxc.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Container represents a container with some metadata
type Container struct {
	ct       *lxc.Container
	Manifest Manifest
}

// NewContainer returns a container struct
func NewContainer(name string) (*Container, error) {
	ct, err := lxc.NewContainer(name)
	if err != nil {
		return nil, err
	}
	return &Container{
		ct: ct,
	}, nil
}

// Create creates new container by clonin parent
func (c *Container) Create(parent string) error {
	orig, err := lxc.NewContainer(parent)
	if err != nil {
		return err
	}
	if err := orig.Clone(c.ct.Name(), lxc.CloneOptions{}); err != nil {
		return err
	}
	ct, err := lxc.NewContainer(c.ct.Name())
	if err != nil {
		return err
	}
	c.ct = ct
	return nil
}

// Stop stops the container
func (c *Container) Stop() error {
	return c.ct.Stop()
}

// Destroy destroys the container
func (c *Container) Destroy() error {
	return c.ct.Destroy()
}

// Start starts the container and wait for IP allocation
func (c *Container) Start() error {
	if err := c.ct.Start(); err != nil {
		return err
	}
	if _, err := c.ct.WaitIPAddresses(30 * time.Second); err != nil {
		log.Errorf("Failed to while waiting to start the container %s. Error: %v", c.ct.Name(), err)
		return err
	}
	return nil
}

// UpdateUTS changes the container's nameand rootfs path
func (c *Container) UpdateUTS(name string) error {
	rootfs := filepath.Join(lxc.GlobalConfigItem("lxc.lxcpath"), name, "rootfs")
	if err := c.ct.SetConfigItem("lxc.utsname", name); err != nil {
		return err
	}
	if err := c.ct.SetConfigItem("lxc.rootfs", rootfs); err != nil {
		return err
	}
	return c.ct.SaveConfigFile(c.ct.ConfigFileName())
}

// RunCommand runs a command inside the container with enviroment, workdir, user as specified
// by its manifest
func (c *Container) RunCommand(command []string) error {
	options := lxc.DefaultAttachOptions
	options.Cwd = "/root"
	options.Env = MinimalEnv
	log.Debugf("Exec environment: %#v\n", options.Env)
	rootfs := c.ct.ConfigItem("lxc.rootfs")[0]
	var buffer bytes.Buffer
	buffer.WriteString("#!/bin/bash\n")
	for _, v := range c.Manifest.Env {
		if _, err := buffer.WriteString("export " + v + "\n"); err != nil {
			return err
		}
	}
	options.ClearEnv = true
	if c.Manifest.WorkDir != "" {
		buffer.WriteString("cd " + c.Manifest.WorkDir + "\n")
	}
	if c.Manifest.User != "" {
		buffer.WriteString("su - " + c.Manifest.User + "\n")
	}
	buffer.WriteString(strings.Join(command, " "))
	file := filepath.Join(rootfs, "/tmp/dockerfile.sh")
	err := ioutil.WriteFile(file, buffer.Bytes(), 0755)
	if err != nil {
		log.Errorf("Failed to open file %s. Error: %v", file, err)
		return err
	}

	exitCode, err := c.ct.RunCommandStatus([]string{"/bin/bash", "/tmp/dockerfile.sh"}, options)
	if err != nil {
		log.Errorf("Failed to execute command: '%s'. Error: %v", command, err)
		return err
	}
	if exitCode != 0 {
		log.Warnf("Failed to execute command: '%s'. Exit code: %d", strings.Join(command, " "), exitCode)
		return fmt.Errorf("Failed to execute command: '%s'. Exit code: %d", strings.Join(command, " "), exitCode)
	}
	return nil
}

// BindMount sets up bind mount for the container, where the input string
// specifies the host directory, container directory and mount options
// separated by ":"
func (c *Container) BindMount(volume string) error {
	parts := strings.Split(volume, ":")
	options := []string{"none", "bind,create=dir", "0", "0"}
	var hostDir string
	var containerDir string
	switch len(parts) {
	case 1:
		containerDir = volume
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		hostDir = dir
	case 2:
		containerDir = parts[1]
		p, err := filepath.Abs(parts[0])
		if err != nil {
			return err
		}
		hostDir = p
	case 3:
		containerDir = parts[1]
		p, err := filepath.Abs(parts[0])
		if err != nil {
			return err
		}
		hostDir = p
		options[1] = "bind," + parts[2]
	default:
		return fmt.Errorf("Invalid volume spec. Parts: %d", len(parts))
	}
	containerDir = strings.TrimPrefix(containerDir, "/")
	val := hostDir + " " + containerDir + " " + strings.Join(options, " ")
	path := c.ct.ConfigFileName()
	if err := c.ct.SetConfigItem("lxc.mount.entry", val); err != nil {
		return err
	}
	return c.ct.SaveConfigFile(path)
}
