package container

import (
	"fmt"
	"gopkg.in/lxc/go-lxc.v2"
	"os"
	"testing"
)

func setup() error {
	ct, err := lxc.NewContainer("trusty")
	if err != nil {
		return err
	}
	options := lxc.TemplateOptions{
		Template: "download",
		Distro:   "ubuntu",
		Release:  "trusty",
		Arch:     "amd64",
	}
	if !ct.Defined() {
		if err := ct.Create(options); err != nil {
			return err
		}
	}
	return nil
}

func teardown() error {
	fmt.Println("Main teardown..")
	ct, err := lxc.NewContainer("trusty")
	if err != nil {
		return err
	}
	if ct.Defined() {
		if err := ct.Destroy(); err != nil {
			return err
		}
	}
	return nil
}

func TestMain(m *testing.M) {
	var code int
	if err := setup(); err != nil {
		code = 1
	}
	code = m.Run()
	if err := teardown(); err != nil {
		code = 1
	}
	os.Exit(code)
}

func Test_Container(t *testing.T) {
	fmt.Println("Container tests...")
	ct, err := NewContainer("nut-test-container")
	if err != nil {
		t.Fatal(err)
	}
	if err := ct.Create("trusty"); err != nil {
		t.Fatal(err)
	}
	if err := ct.Start(); err != nil {
		t.Fatal(err)
	}
	if err := ct.RunCommand([]string{"apt-get", "update", "-y"}); err != nil {
		t.Fatal(err)
	}
	if err := ct.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := ct.Destroy(); err != nil {
		t.Fatal(err)
	}
}
