package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/adapters/atlassian"
	"github.com/masonhuemmer/atlas/internal/adapters/keychain"
	"github.com/masonhuemmer/atlas/internal/app/auth"
	"github.com/masonhuemmer/atlas/internal/config"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func testDeps() (Deps, *bytes.Buffer, *bytes.Buffer) {
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	d := Deps{
		Config: config.Config{},
		Store:  &keychain.Fake{},
		Login:  auth.FakeLogin(),
		Jira:   atlassian.JiraAPI{Memory: atlassian.Seed()},
		Stdout: out,
		Stderr: errw,
	}
	return d, out, errw
}

func TestHelpNoSession(t *testing.T) {
	d, out, _ := testDeps()
	code := Run([]string{"atlas", "--help"}, d)
	if code != 0 {
		t.Fatal(code)
	}
	s := out.String()
	for _, want := range []string{"auth", "site", "jira", "mcp", "JSON", "3", "4", "5", "6"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}

func TestJSONHumanMutex(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "--json", "--human", "auth", "status"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw)
	}
}
