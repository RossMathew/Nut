package container

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/lxc/go-lxc.v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Image represent a container image, which holds rootfs and metadata
type Image struct {
	Path string
	ct   *lxc.Container
}

// NewImage Returns a Image struct for the provided container name and
// image path
func NewImage(name, path string) (*Image, error) {
	ct, err := lxc.NewContainer(name)
	if err != nil {
		return nil, err
	}
	return &Image{ct: ct, Path: path}, nil
}

// Create creates a new tarball image from a container.
// sudo is used for invoking tar, if set to true
func (i *Image) Create(sudo bool) error {
	//ExportContainer(string, string, bool) error
	lxcdir := lxc.GlobalConfigItem("lxc.lxcpath")
	ctDir := filepath.Join(lxcdir, i.ct.Name())
	command := fmt.Sprintf("tar -Jcpf %s --numeric-owner -C %s .", i.Path, ctDir)
	if sudo {
		command = "sudo " + command
	}
	parts := strings.Fields(command)
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(out))
		log.Error(err)
		return err
	}
	return nil
}

// Decompress decompress the image into a container
func (i *Image) Decompress(sudo bool) error {
	lxcpath := lxc.GlobalConfigItem("lxc.lxcpath")
	ctDir := filepath.Join(lxcpath, i.ct.Name())
	untarCommand := fmt.Sprintf("tar --numeric-owner -xpJf  %s -C %s", i.Path, ctDir)
	if sudo {
		untarCommand = "sudo " + untarCommand
	}
	if err := os.Mkdir(ctDir, 0770); err != nil {
		log.Errorln(err)
		return err
	}
	log.Infof("Invoking: %s", untarCommand)
	parts := strings.Fields(untarCommand)
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(out))
		log.Error(err)
		return err
	}
	return nil
}

// Publish publishes the image in s3
func (i *Image) Publish(region, bucket, key string) error {
	svc := s3.New(session.New(), aws.NewConfig().WithRegion("region"))
	fi, err := os.Open(i.Path)
	if err != nil {
		return err
	}
	defer fi.Close()
	params := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   fi,
	}
	_, uploadErr := svc.PutObject(params)
	return uploadErr
}
