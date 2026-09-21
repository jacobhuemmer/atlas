package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var forbidden = []string{
	"sesamidevel",
	"sesami-io",
	"sesamiio",
	"gardaworld",
	"garda",
	"jsm-garda",
}

func TestProductionSourceHasNoTenantNames(t *testing.T) {
	root := moduleRoot(t)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == ".git" || base == ".humanlayer" || base == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		name := info.Name()
		if strings.HasSuffix(name, "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if !keep(rel) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		low := strings.ToLower(string(b))
		for _, bad := range forbidden {
			if strings.Contains(low, strings.ToLower(bad)) {
				t.Errorf("%s contains %q", rel, bad)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func keep(rel string) bool {
	switch {
	case strings.HasPrefix(rel, "cmd/"),
		strings.HasPrefix(rel, "internal/"),
		strings.HasPrefix(rel, "Formula/"),
		strings.HasPrefix(rel, "specs/"),
		strings.HasPrefix(rel, "features/"):
		return true
	default:
		return false
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}
