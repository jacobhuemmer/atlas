package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestLoadFakeThreeCloudFixture(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	cfg, err := LoadFake()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sites) != 3 {
		t.Fatalf("%+v", cfg.Sites)
	}
	s, err := domain.Lookup("sesamidevel")
	if err != nil || s.Hostname != "sesamidevel.atlassian.net" {
		t.Fatalf("%+v %v", s, err)
	}
	if domain.DefaultWorkspace() != "sesamiio" {
		t.Fatal(domain.DefaultWorkspace())
	}
	if domain.DefaultJSMSite() != "garda" {
		t.Fatal(domain.DefaultJSMSite())
	}
}

func TestLoadATLASConfigWins(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	dir := t.TempDir()
	path := filepath.Join(dir, "atlas.toml")
	body := `
[defaults]
workspace = "fromfile"
[[sites]]
alias = "dev"
hostname = "dev.example.atlassian.net"
uuid = "00000000-0000-0000-0000-000000000001"
role = "licensed"
[projects]
ABC = "dev"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATLAS_CONFIG", path)
	t.Setenv("ATLAS_FAKE", "")
	t.Setenv("ATLAS_WORKSPACE", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sites) != 1 || cfg.Sites[0].Alias != "dev" {
		t.Fatalf("%+v", cfg.Sites)
	}
	if domain.DefaultWorkspace() != "fromfile" {
		t.Fatal(domain.DefaultWorkspace())
	}
}

func TestLoadEnvWorkspaceOverlay(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	dir := t.TempDir()
	path := filepath.Join(dir, "atlas.toml")
	body := `
[defaults]
workspace = "fromfile"
[[sites]]
alias = "dev"
hostname = "dev.example.atlassian.net"
role = "licensed"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATLAS_CONFIG", path)
	t.Setenv("ATLAS_FAKE", "")
	t.Setenv("ATLAS_WORKSPACE", "fromenv")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Defaults.Workspace != "fromenv" {
		t.Fatalf("%+v", cfg.Defaults)
	}
}

func TestLoadMissingIsUsage(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	t.Setenv("ATLAS_CONFIG", "")
	t.Setenv("ATLAS_FAKE", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	_, err := Load()
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	if err == nil || err.Error() == "" {
		t.Fatal("expected message")
	}
}

func TestLoadUnknownKeyIsUsage(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("[defaults]\nnope = \"x\"\n[[sites]]\nalias=\"a\"\nhostname=\"h\"\nrole=\"licensed\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATLAS_CONFIG", path)
	t.Setenv("ATLAS_FAKE", "")
	_, err := Load()
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestLoadDuplicateAliasIsUsage(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	dir := t.TempDir()
	path := filepath.Join(dir, "dup.toml")
	body := `
[[sites]]
alias = "dev"
hostname = "a.example.atlassian.net"
role = "licensed"
[[sites]]
alias = "dev"
hostname = "b.example.atlassian.net"
role = "licensed"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATLAS_CONFIG", path)
	t.Setenv("ATLAS_FAKE", "")
	_, err := Load()
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}

func TestLoadJSON(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	body := `{
  "defaults": {"workspace": "ws", "jsm_site": "helpdesk"},
  "sites": [
    {"alias": "dev", "hostname": "dev.example.atlassian.net", "role": "licensed"},
    {"alias": "helpdesk", "hostname": "helpdesk.example.atlassian.net", "role": "jsm_customer"}
  ],
  "projects": {"ABC": "dev"},
  "spaces": {"DOCS": "dev"}
}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATLAS_CONFIG", path)
	t.Setenv("ATLAS_FAKE", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Defaults.JSMSite != "helpdesk" || cfg.Projects["ABC"] != "dev" {
		t.Fatalf("%+v", cfg)
	}
}

func TestLoadXDGToml(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	xdg := t.TempDir()
	dir := filepath.Join(xdg, "atlas")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := `
[[sites]]
alias = "only"
hostname = "only.example.atlassian.net"
role = "licensed"
`
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATLAS_CONFIG", "")
	t.Setenv("ATLAS_FAKE", "")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sites) != 1 || cfg.Sites[0].Alias != "only" {
		t.Fatalf("%+v", cfg.Sites)
	}
}
