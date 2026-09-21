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
