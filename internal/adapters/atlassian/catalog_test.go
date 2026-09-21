package atlassian

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

func host(alias string) string {
	s, err := domain.Lookup(alias)
	if err != nil {
		panic(err)
	}
	return s.Hostname
}
