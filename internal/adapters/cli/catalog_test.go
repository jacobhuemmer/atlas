package cli

import (
	"os"
	"testing"

	"github.com/masonhuemmer/atlas/internal/config"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMain(m *testing.M) {
	if _, err := config.LoadFake(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func fixtureHost(t *testing.T, alias string) string {
	t.Helper()
	s, err := domain.Lookup(alias)
	if err != nil {
		t.Fatal(err)
	}
	return s.Hostname
}

func jsmHost(t *testing.T) string {
	t.Helper()
	alias := domain.DefaultJSMSite()
	if alias == "" {
		t.Fatal("fixture jsm_site empty")
	}
	return fixtureHost(t, alias)
}
