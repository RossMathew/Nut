package container

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
)

func (c *Container) addFiles(src, dest string) error {
	rootfs := c.ct.ConfigItem("lxc.rootfs")[0]
	base := filepath.Base(src)
	tmpContainer := filepath.Join(rootfs, "tmp", base)
	cmd := exec.Command("/bin/cp", "-ar", src, tmpContainer)
	log.Warnln("/bin/cp", "-ar", src, tmpContainer)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Errorln("Failed to copy temporary files from host to container tmp directory")
		log.Errorln("Error:", err)
		log.Errorln("Output:", out)
		return err
	}
	if err := c.RunCommand([]string{"cp", "-r", filepath.Join("/tmp", base), dest}); err != nil {
		log.Errorf("Failed to copy temporary files within container's /tmp to target directory. Error: %s\n", err)
		return err
	}
	rmCmd := exec.Command("/bin/rm", "-rf", tmpContainer)
	if err := rmCmd.Run(); err != nil {
		log.Error("Failed to delete temporary files")
		return err
	}
	return nil
}

func (c *Container) fetchArtifacts() error {
	rootfs := c.ct.ConfigItem("lxc.rootfs")[0]
	for k, v := range c.Manifest.Labels {
		if strings.HasPrefix(k, "nut_artifact_") {
			artifact := filepath.Base(v)
			if err := c.RunCommand([]string{"cp", "-r", v, filepath.Join("/tmp", artifact)}); err != nil {
				log.Errorf("Failed to copy artifact to /tmp. Error: %s\n", err)
				return err
			}
			pathInContainer := filepath.Join(rootfs, "tmp", artifact)
			cmd := exec.Command("/bin/cp", "-ar", pathInContainer, artifact)
			if err := cmd.Run(); err != nil {
				log.Errorf("Failed to copy files from container to host. Error: %s\n", err)
			}
		}
	}
	return nil
}

func (c *Container) writeManifest() error {
	rootfs := c.ct.ConfigItem("lxc.rootfs")[0]
	manifestPath := filepath.Join(rootfs, "../manifest.yml")
	d, err := yaml.Marshal(&c.Manifest)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(manifestPath, d, 0644)
}
