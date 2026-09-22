package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestAuthStatusSignedOutLoginLogout(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "auth", "status"}, d)
	if code != 0 {
		t.Fatal(errw.String())
	}
	var st map[string]any
	if err := json.Unmarshal(out.Bytes(), &st); err != nil {
		t.Fatal(err, out.String())
	}
	if st["signed_in"] != false {
		t.Fatal(st)
	}
	if strings.Contains(strings.ToLower(out.String()), "token") {
		t.Fatal("token leaked")
	}
	sites, _ := st["sites"].([]any)
	if len(sites) != 3 {
		t.Fatalf("sites %v", sites)
	}
	out.Reset()
	code = Run([]string{"atlas", "auth", "login", "--site", "sesamidevel", "--email", "a@b.c", "--token", "secret-token"}, d)
	if code != 0 {
		t.Fatal(errw.String())
	}
	if strings.Contains(out.String(), "secret-token") {
		t.Fatal(out.String())
	}
	out.Reset()
	_ = Run([]string{"atlas", "auth", "status"}, d)
	_ = json.Unmarshal(out.Bytes(), &st)
	if st["signed_in"] != true {
		t.Fatal(st)
	}
	if strings.Contains(out.String(), "secret-token") {
		t.Fatal(out.String())
	}
	out.Reset()
	if c := Run([]string{"atlas", "auth", "logout"}, d); c != 0 {
		t.Fatal(c)
	}
	out.Reset()
	_ = Run([]string{"atlas", "auth", "status"}, d)
	_ = json.Unmarshal(out.Bytes(), &st)
	if st["signed_in"] != false {
		t.Fatal(st)
	}
}

func TestAuthLoginFromOpIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "auth", "login", "--site", "sesamidevel", "--from-op"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
}

func TestAuthWorkspaceLoginStatusAndTargetedLogout(t *testing.T) {
	d, out, errw := testDeps()
	if code := Run([]string{"atlas", "auth", "login", "--site", "sesamidevel", "--email", "site@example.com", "--token", "site-secret"}, d); code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	out.Reset()
	errw.Reset()
	if code := Run([]string{"atlas", "auth", "login", "--workspace", domain.DefaultWorkspace(), "--email", "bitbucket@example.com", "--token", "workspace-secret"}, d); code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	if strings.Contains(out.String(), "site-secret") || strings.Contains(out.String(), "workspace-secret") {
		t.Fatal("credential leaked", out.String())
	}

	var st struct {
		SignedIn      bool `json:"signed_in"`
		SessionUsable bool `json:"session_usable"`
		Sites         []struct {
			Alias  string `json:"alias"`
			Usable bool   `json:"usable"`
		} `json:"sites"`
		Workspaces []struct {
			Slug   string `json:"slug"`
			Usable bool   `json:"usable"`
		} `json:"workspaces"`
	}
	if err := json.Unmarshal(out.Bytes(), &st); err != nil {
		t.Fatal(err, out.String())
	}
	if !st.SignedIn || !st.SessionUsable || !usableSite(st.Sites, "sesamidevel") || !usableWorkspace(st.Workspaces, domain.DefaultWorkspace()) {
		t.Fatalf("unexpected status: %+v", st)
	}

	out.Reset()
	errw.Reset()
	if code := Run([]string{"atlas", "auth", "logout", "--workspace", domain.DefaultWorkspace()}, d); code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	if err := json.Unmarshal(out.Bytes(), &st); err != nil {
		t.Fatal(err, out.String())
	}
	if !usableSite(st.Sites, "sesamidevel") || usableWorkspace(st.Workspaces, domain.DefaultWorkspace()) {
		t.Fatalf("targeted logout removed the wrong credential: %+v", st)
	}
}

func TestAuthLoginAndLogoutRejectSiteWorkspaceConflict(t *testing.T) {
	for _, args := range [][]string{
		{"atlas", "auth", "login", "--site", "sesamidevel", "--workspace", domain.DefaultWorkspace(), "--email", "a@b.c", "--token", "secret"},
		{"atlas", "auth", "logout", "--site", "sesamidevel", "--workspace", domain.DefaultWorkspace()},
	} {
		d, _, errw := testDeps()
		if code := Run(args, d); code != domain.ExitUsage {
			t.Fatalf("%v: code=%d stderr=%s", args, code, errw.String())
		}
	}
}

func usableSite(sites []struct {
	Alias  string `json:"alias"`
	Usable bool   `json:"usable"`
}, alias string) bool {
	for _, site := range sites {
		if site.Alias == alias {
			return site.Usable
		}
	}
	return false
}

func usableWorkspace(workspaces []struct {
	Slug   string `json:"slug"`
	Usable bool   `json:"usable"`
}, slug string) bool {
	for _, workspace := range workspaces {
		if workspace.Slug == slug {
			return workspace.Usable
		}
	}
	return false
}

func TestSiteListAndResolve(t *testing.T) {
	d, out, errw := testDeps()
	if c := Run([]string{"atlas", "site", "list"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	s := out.String()
	for _, want := range []string{"sesamidevel", "sesami-io", "garda", "licensed", "jsm_customer"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
	out.Reset()
	if c := Run([]string{"atlas", "site", "resolve", "CAB-not-a-site"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	out.Reset()
	if c := Run([]string{"atlas", "site", "resolve", "sesamidevel.atlassian.net"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	if !strings.Contains(out.String(), "sesamidevel") {
		t.Fatal(out.String())
	}
}
