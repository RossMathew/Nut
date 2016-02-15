package container

import (
	"os"
	"testing"
)

func Test_Image(t *testing.T) {
	i1, err := NewImage("trusty", "trusty.tgz")
	if err != nil {
		t.Fatal(err)
	}
	if err := i1.Create(true); err != nil {
		t.Fatal(err)
	}
	i2, err := NewImage("trusty1", "trusty.tgz")
	if err != nil {
		t.Fatal(err)
	}
	if err := i2.Decompress(true); err != nil {
		t.Fatal(err)
	}
	os.Remove("trusty.tgz")
	ct, err := NewContainer("trusty1")
	if err != nil {
		t.Fatal(err)
	}
	if err := ct.UpdateUTS("trusty1"); err != nil {
		t.Fatal(err)
	}

	if err := ct.Destroy(); err != nil {
		t.Fatal(err)
	}
}
