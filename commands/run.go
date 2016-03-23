package commands

import (
	"flag"
	"fmt"
	"github.com/PagerDuty/nut/container"
	"github.com/mitchellh/cli"
	log "github.com/sirupsen/logrus"
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
	return strings.TrimSpace(helpText) + AddCommonHelp()
}

func (command *RunCommand) Synopsis() string {
	return "Run command/entrypoint inside a container"
}

func (command *RunCommand) Run(args []string) int {
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	flagSet.Usage = func() { fmt.Println(command.Help()) }
	cmd := flagSet.String("command", "", "Command to run inside the container")
	AddCommonFlags(flagSet)
	if err := flagSet.Parse(args); err != nil {
		log.Errorln(err)
		return 1
	}
	ConfigureLogging()
	if len(flagSet.Args()) != 1 {
		log.Errorln("You have to pass container name as argument")
		return 1
	}
	name := flagSet.Args()[0]
	ct, err := container.NewContainer(name)
	if err != nil {
		log.Errorln(err)
		return 1
	}
	if err := ct.Manifest.Load(name); err != nil {
		log.Errorln("Failed to load container manifest")
		log.Errorln(err)
		return 1
	}
	cmdParts := ct.Manifest.EntryPoint
	if *cmd != "" {
		cmdParts = strings.Fields(*cmd)
	}
	if err := ct.RunCommand(cmdParts); err != nil {
		log.Errorln(err)
		return 1
	}
	return 0
}
