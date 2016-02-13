package commands

import (
	"flag"
	"fmt"
	"github.com/PagerDuty/nut/specification"
	"github.com/mitchellh/cli"
	log "github.com/sirupsen/logrus"
	"gopkg.in/lxc/go-lxc.v2"
	"strings"
)

type RunCommand struct{}

func Run() (cli.Command, error) {
	command := &RunCommand{}
	return command, nil
}

func (command *RunCommand) Help() string {
	helpText := `
	Usage: nut run [options] name

		Run entrypoint or command inside a container

  Options:
		-command Command to run (quoted)
	`
	return strings.TrimSpace(helpText)
}

func (command *RunCommand) Synopsis() string {
	return "Run command/entrypoint inside a container"
}

func (command *RunCommand) Run(args []string) int {
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	flagSet.Usage = func() { fmt.Println(command.Help()) }
	cmd := flagSet.String("command", "", "Command to run inside the container")
	if err := flagSet.Parse(args); err != nil {
		log.Errorln(err)
		return 1
	}
	if len(flagSet.Args()) != 1 {
		log.Errorln("You have to pass container name as argument")
		return 1
	}
	name := flagSet.Args()[0]
	spec := specification.New(name)
	ct, err := lxc.NewContainer(name)
	if err != nil {
		log.Errorf("Failed to initialize container object. Error: %v", err)
		return 1
	}
	spec.State.Container = ct
	if err := spec.State.Manifest.Load(name); err != nil {
		log.Warnln("Failed to load manifest from patent container. Error:", err)
	} else {
		spec.State.Env = spec.State.Manifest.Env
		spec.State.Cwd = spec.State.Manifest.WorkDir
	}
	cmdParts := spec.State.Manifest.EntryPoint
	if *cmd != "" {
		cmdParts = strings.Fields(*cmd)
	}
	if err := spec.RunCommand(cmdParts); err != nil {
		log.Errorln(err)
		return 1
	}
	return 0
}
