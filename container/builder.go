package container

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Builder represents a container build environment
type Builder struct {
	Name       string
	Volumes    []string
	Statements []string
	RootDir    string
}

// NewBuilder returns a Builder struct
func NewBuilder(name string) *Builder {
	return &Builder{
		Name: name,
	}
}

// Parse take a dockerfile like DSL file path and populates build instructions
func (b *Builder) Parse(file string) error {
	fi, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fi.Close()
	var statements []string
	scanner := bufio.NewScanner(fi)
	scanner.Split(bufio.ScanLines)
	var isComment = regexp.MustCompile(`^#`)
	var isExtendedStatement = regexp.MustCompile(`\\$`)
	previousStatement := ""
	for scanner.Scan() {
		line := scanner.Text()
		if isComment.MatchString(line) {
			continue
		} else if isExtendedStatement.MatchString(line) {
			// if line ends with \ then append statement
			if previousStatement != "" {
				previousStatement = previousStatement + " " + strings.TrimRight(line, "\\")
			} else {
				previousStatement = strings.TrimRight(line, "\\")
			}
		} else if strings.TrimSpace(line) == "" {
			// dont process if line empty
			continue
		} else {
			// if line does not end with \ then append statement
			var statement string
			if previousStatement != "" {
				statement = previousStatement + " " + line
				previousStatement = ""
			} else {
				statement = line
			}
			statements = append(statements, statement)
		}
	}
	b.Statements = statements
	absPath, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	b.RootDir = filepath.Dir(absPath)
	return nil
}

func (b *Builder) createContainer(from string) (*Container, error) {
	parent := TagToName(from)
	c, err := NewContainer(b.Name)
	if err != nil {
		return nil, err
	}
	if err := c.Create(parent); err != nil {
		return nil, err
	}
	log.Infoln("Created container named ", b.Name)
	for _, volume := range b.Volumes {
		if err = c.BindMount(volume); err != nil {
			return nil, err
		}
	}
	if err := c.Start(); err != nil {
		return nil, err
	}
	if err = c.Manifest.Load(parent); err != nil {
		log.Warnf("Failed to load manifest from patent container. Error: %s\n", err)
	}
	return c, nil
}

// Build creates a new container from  build instructions and return the container
// struct
func (b *Builder) Build() (*Container, error) {
	var c *Container
	var err error
	for _, statement := range b.Statements {
		words := strings.Fields(statement)
		switch words[0] {
		case "FROM":
			if c != nil {
				return nil, errors.New("Container already built. Multiple FROM declaration?")
			}
			c, err = b.createContainer(words[1])
			if err != nil {
				return nil, err
			}
		case "RUN":
			if c == nil {
				log.Error("No container has been created yet. Use FROM directive")
				return nil, errors.New("No container has been created yet. Use FROM directive")
			}
			command := words[1:len(words)]
			if err := c.RunCommand(command); err != nil {
				log.Errorf("Failed to run command inside container. Error: %s\n", err)
				return nil, err
			}
		case "ENV":
			for i := 1; i < len(words); i++ {
				if strings.Contains(words[i], "=") {
					c.Manifest.Env = append(c.Manifest.Env, words[i])
				} else {
					c.Manifest.Env = append(c.Manifest.Env, words[i]+"="+words[i+1])
					i++
				}
			}
		case "WORKDIR":
			c.Manifest.WorkDir = words[1]
		case "ADD":
			if err := c.addFiles(filepath.Join(b.RootDir, words[1]), words[2]); err != nil {
				return nil, err
			}
		case "COPY":
			if err := c.addFiles(filepath.Join(b.RootDir, words[1]), words[2]); err != nil {
				return nil, err
			}
		case "LABEL":
			for i := 1; i < len(words); i++ {
				if strings.Contains(words[i], "=") {
					pair := strings.Split(words[i], "=")
					c.Manifest.Labels[pair[0]] = pair[1]
				} else {
					return nil, errors.New("Invalid LABEL instruction. LABELS must have '=' in them")
				}
			}
		case "EXPOSE":
			for _, p := range words[1:len(words)] {
				port, err := strconv.ParseUint(p, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Error parsing ports in EXPOSE instruction. Err:%s\n", err)
				}
				c.Manifest.ExposedPorts = append(c.Manifest.ExposedPorts, port)
			}
		case "MAINTAINER":
			c.Manifest.Maintainers = append(c.Manifest.Maintainers, strings.Join(words[1:len(words)], " "))
		case "USER":
			c.Manifest.User = words[1]
		case "VOLUME":
			// FIXME
		case "STOPSIGNAL":
			// FIXME
		case "CMD":
			c.Manifest.EntryPoint = words[1:]
		case "ENTRYPOINT":
			c.Manifest.EntryPoint = words[1:]
		default:
			return nil, fmt.Errorf("Unknown instruction: %s", words[0])
		}
	}
	if err = c.fetchArtifacts(); err != nil {
		return c, err
	}
	return c, c.writeManifest()
}
