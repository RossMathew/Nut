package commands

import (
	"flag"
	"fmt"
	"github.com/PagerDuty/nut/specification"
	"github.com/mitchellh/cli"
	log "github.com/sirupsen/logrus"
	"strings"
)

type BuildCommand struct {
}

func Build() (cli.Command, error) {
	command := &BuildCommand{}
	return command, nil
}

func (command *BuildCommand) Help() string {
	helpText := `
		-specfile    Local path to the specification file (defaults to dockerfle)
		-ephemeral   Destroy the container after creation
		-name        Name of the container (defaults to randomly generated UUID)
		-volume      Mount host directory inside container
	`
	return strings.TrimSpace(helpText)
}

func (command *BuildCommand) Synopsis() string {
	synopsis := "Build container from Dockerfile"
	return synopsis
}

func (command *BuildCommand) Run(args []string) int {

	flagSet := flag.NewFlagSet("build", flag.ExitOnError)
	flagSet.Usage = func() { fmt.Println(command.Help()) }

	file := flagSet.String("specfile", "Dockerfile", "Container build specification file")
	ephemeral := flagSet.Bool("ephemeral", false, "Destroy the container after creating it")
	name := flagSet.String("name", "", "Name of the resulting container (defaults to randomly generated UUID)")
	volume := flagSet.String("volume", "", "Mount host directory inside container. Format: '[host_directory:]container_directory[:mount options]")

	flagSet.Parse(args)
	if *name == "" {
		uuid, err := specification.UUID()
		if err != nil {
			log.Errorln(err)
			return -1
		}
		name = &uuid
	}

	spec := specification.New(*name)

	if err := spec.Parse(*file); err != nil {
		log.Errorf("Failed to parse dockerfile. Error: %s\n", err)
		return -1
	}

	if err := spec.Build(*volume); err != nil {
		log.Errorf("Failed to build container from dockerfile. Error: %s\n", err)
		return -1
	}

	if *ephemeral {
		log.Infof("Ephemeral mode. Destroying the container")
		if err := spec.Destroy(); err != nil {
			log.Errorf("Failed to destroy container. Error: %s\n", err)
			return -1
		}
	}
	return 0
}
