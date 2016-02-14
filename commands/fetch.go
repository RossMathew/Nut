package commands

import (
	"flag"
	"fmt"
	"github.com/PagerDuty/nut/container"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/cli"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type FetchCommand struct {
}

func Fetch() (cli.Command, error) {
	command := &FetchCommand{}
	return command, nil
}

func (command *FetchCommand) Help() string {
	helpText := `
		-region   S3 region
		-bucket   S3 bucket name
		-key      S3 key name
		-name     Name of the container (Default: random uuid)
		-sudo     Use sudo to decompress image
	`
	return strings.TrimSpace(helpText)
}

func (command *FetchCommand) Synopsis() string {
	return "Create container from rootfs tarball stored in s3"
}

func (command *FetchCommand) Run(args []string) int {
	flagSet := flag.NewFlagSet("fetch", flag.ExitOnError)
	flagSet.Usage = func() { fmt.Println(command.Help()) }
	name := flagSet.String("name", "", "Name of the container (Default: random UUID)")
	bucket := flagSet.String("bucket", "", "S3 bucket")
	key := flagSet.String("key", "", "S3 key")
	region := flagSet.String("region", "us-west-1", "S3 region")
	sudo := flagSet.Bool("sudo", false, "Use sudo during decompression of image")
	flagSet.Parse(args)
	if *bucket == "" {
		log.Errorf("Must provide the s3 bucket name")
		return -1
	}
	if *key == "" {
		log.Errorf("Must provide the s3 key name")
		return -1
	}
	svc := s3.New(session.New(), aws.NewConfig().WithRegion(*region))
	fo, err := ioutil.TempFile(os.TempDir(), "nut")
	if err != nil {
		log.Error(err)
		return -1
	}
	params := &s3.GetObjectInput{
		Bucket: aws.String(*bucket),
		Key:    aws.String(*key),
	}

	resp, downloadErr := svc.GetObject(params)
	if downloadErr != nil {
		log.Error(downloadErr)
		return -1
	}
	defer resp.Body.Close()
	if _, copyError := io.Copy(fo, resp.Body); copyError != nil {
		log.Errorln(copyError)
		return -1
	}
	log.Infof("Image written to: %s\n", fo.Name())
	if *name == "" {
		uuid, err := container.UUID()
		if err != nil {
			log.Errorln(err)
			return -1
		}
		name = &uuid
	}
	i, err := container.NewImage(*name, fo.Name())
	if err != nil {
		log.Errorln(err)
		return -1
	}
	fo.Close()
	if err := i.Decompress(*sudo); err != nil {
		log.Errorln(err)
		return -1
	}
	ct, err := container.NewContainer(*name)
	if err != nil {
		log.Errorln(err)
		return -1
	}
	if err := ct.UpdateUTS(*name); err != nil {
		log.Errorln(err)
		return -1
	}
	return 0
}
