package commands

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	logfd    *os.File
	loglevel *string
	logfile  *string
)

func AddCommonFlags(flagSet *flag.FlagSet) {
	logfile = flagSet.String("log-output", "", "Log output file")
	loglevel = flagSet.String("log-level", "", "Set Log level")
}

func ConfigureLogging() {
	switch *loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.Warn("Unknow log level", *loglevel, "falling back to info level")
		log.SetLevel(log.InfoLevel)
	}
	if *logfile == "" {
		return
	}
	f, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Errorln("Failed to open log file", *logfile)
		log.Fatal(err)
	}
	logfd = f
	log.SetOutput(f)
}

func AddCommonHelp() string {
	helpText := `
		-log-output=output    Path of log file
		-log-level=info       Set log level
`
	return helpText
}
