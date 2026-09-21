package auth

import (
	"os"
	"testing"

	"github.com/masonhuemmer/atlas/internal/config"
)

func TestMain(m *testing.M) {
	if _, err := config.LoadFake(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
