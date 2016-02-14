package container

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

var (
	// MinimalEnv used to set exec environment for command invocation inside lxc
	MinimalEnv = []string{
		"SHELL=/bin/bash",
		"USER=root",
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"PWD=/root",
		"EDITOR=vim",
		"LANG=en_US.UTF-8",
		"HOME=/root",
		"LANGUAGE=en_US",
		"LOGNAME=root",
	}
)

// UUID generates uuid
func UUID() (string, error) {
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		return "", err
	}
	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F
	return hex.EncodeToString(u), nil
}

// TagToName converts Dockerfile's FROM target to a valid lxc container name
func TagToName(tag string) string {
	// docker FROM entry: org/repo:version -> org-repo_version
	orgConverted := strings.Replace(tag, "/", "-", 1)
	repoConverted := strings.Replace(orgConverted, ":", "_", 1)
	return repoConverted
}
