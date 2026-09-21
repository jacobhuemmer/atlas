package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/config"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestSiteListFromFakeCatalog(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "site", "list"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var sites []domain.Site
	if err := json.Unmarshal(out.Bytes(), &sites); err != nil {
		t.Fatal(err, out.String())
	}
	got := map[string]bool{}
	for _, s := range sites {
		got[s.Alias] = true
	}
	for _, want := range []string{"sesamidevel", "sesami-io", "garda"} {
		if !got[want] {
			t.Fatalf("missing %s in %+v", want, sites)
		}
	}
	if len(sites) != 3 {
		t.Fatalf("%+v", sites)
	}
}

func TestResolveUsesCatalogMaps(t *testing.T) {
	s, err := domain.Resolve(domain.ResolveInput{Issue: "SDO-1"})
	if err != nil || s.Alias != "sesamidevel" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = domain.Resolve(domain.ResolveInput{Space: "CCAB"})
	if err != nil || s.Alias != "sesami-io" {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestDualAliasJQLStillUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "jira", "search", "--jql", "project in (SDO, CAB)"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
}

func TestJiraOnJSMCustomerIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "jira", "search", "--site", "garda", "--jql", "project = EOS"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	if !strings.Contains(errw.String(), "jsm") {
		t.Fatal(errw.String())
	}
}

func TestConfluenceOnJSMCustomerIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "confluence", "get", "100", "--site", "garda"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
}

func TestPRGetUsesCatalogWorkspace(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "pr", "get", "--repo", "atlas", "--id", "1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var pr domain.PullRequest
	if err := json.Unmarshal(out.Bytes(), &pr); err != nil {
		t.Fatal(err, out.String())
	}
	if pr.Workspace != domain.DefaultWorkspace() {
		t.Fatalf("workspace %q", pr.Workspace)
	}
}

func TestJSMWithoutSiteUsesDefault(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jsm", "desks"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var page domain.DeskList
	if err := json.Unmarshal(out.Bytes(), &page); err != nil {
		t.Fatal(err, out.String())
	}
	if page.Site != jsmHost(t) {
		t.Fatalf("site %q", page.Site)
	}
}

func TestJSMEmptyDefaultRequiresSite(t *testing.T) {
	d, _, errw := testDeps()
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	cat := prev
	cat.Defaults.JSMSite = ""
	domain.SetCatalog(cat)
	code := Run([]string{"atlas", "jsm", "desks"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	if !strings.Contains(errw.String(), "--site") {
		t.Fatal(errw.String())
	}
}

func TestHelpHasNoCustomerNames(t *testing.T) {
	d, out, _ := testDeps()
	code := Run([]string{"atlas", "--help"}, d)
	if code != 0 {
		t.Fatal(code)
	}
	s := strings.ToLower(out.String())
	for _, bad := range []string{"sesami", "garda", "gardaworld", "sesamidevel"} {
		if strings.Contains(s, bad) {
			t.Fatalf("help leaked %q: %s", bad, out.String())
		}
	}
}

func TestMissingConfigIsUsage(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	t.Setenv("ATLAS_CONFIG", "")
	t.Setenv("ATLAS_FAKE", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	_, err := config.Load()
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
	d, _, errw := testDeps()
	code := Fail(d, err)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	if !strings.Contains(errw.String(), "config") {
		t.Fatal(errw.String())
	}
}

func TestATLASConfigSiteListOnlyThoseAliases(t *testing.T) {
	prev := domain.CurrentCatalog()
	t.Cleanup(func() { domain.SetCatalog(prev) })
	dir := t.TempDir()
	path := filepath.Join(dir, "only.toml")
	body := `
[defaults]
workspace = "example"
[[sites]]
alias = "dev"
hostname = "dev.example.atlassian.net"
role = "licensed"
[[sites]]
alias = "helpdesk"
hostname = "helpdesk.example.atlassian.net"
role = "jsm_customer"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ATLAS_CONFIG", path)
	t.Setenv("ATLAS_FAKE", "")
	d, out, errw := testDeps()
	loaded, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	d.Config = loaded
	code := Run([]string{"atlas", "site", "list"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	s := out.String()
	if !strings.Contains(s, "dev") || !strings.Contains(s, "helpdesk") {
		t.Fatal(s)
	}
	if strings.Contains(s, "sesamidevel") || strings.Contains(s, "garda") {
		t.Fatal(s)
	}
}
