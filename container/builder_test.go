package container

import (
	"path/filepath"
	"testing"
)

func Test_Builder(t *testing.T) {
	specfile, err := filepath.Abs("../examples/minimal")
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuilder("nut-test-builder")
	if err := b.Parse(specfile); err != nil {
		t.Fatal(err)
	}
	if len(b.Statements) != 3 {
		t.Fatalf("Number of statements: %d, not 3", len(b.Statements))
	}
	ct, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if ct.Manifest.Maintainers[0] != "foo@example.com" {
		t.Fatal("Maintainer entry not populated")
	}
	if err := ct.Stop(); err != nil {
		t.Fatal("Failed to stop test container")
	}
	if err := ct.Destroy(); err != nil {
		t.Fatal("Failed to destroy test container")
	}
}
