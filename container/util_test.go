package container

import (
	"strings"
	"testing"
)

func Test_UUID(t *testing.T) {
	uuid, err := UUID()
	if err != nil {
		t.Fatalf("Failed to generate uuid, erro: %s", err)
	}
	l := strings.NewReader(uuid).Len()
	if l != 32 {
		t.Fatalf("Length of uuid is %d , not 32", l)
	}
}

func Test_TagToName(t *testing.T) {
	parent := TagToName("golang:1.5")
	if parent != "golang_1.5" {
		t.Fatalf("Expected: golang_1.5, found: %s", parent)
	}
}
