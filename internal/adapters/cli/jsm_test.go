package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/adapters/atlassian"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestJSMDesksSeedNoJiraSearch(t *testing.T) {
	d, out, errw := testDeps()
	mem := d.JSM.(atlassian.JSMAPI).Memory
	before := mem.SearchCalls()
	code := Run([]string{"atlas", "jsm", "desks"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	if mem.SearchCalls() != before {
		t.Fatal("jsm desks must not call Jira issue search")
	}
	var page domain.DeskList
	if err := json.Unmarshal(out.Bytes(), &page); err != nil {
		t.Fatal(err, out.String())
	}
	if page.Site != domain.GardaHostname {
		t.Fatalf("site %q", page.Site)
	}
	got := map[string]string{}
	for _, desk := range page.Items {
		got[desk.Key] = desk.ID
	}
	if got["EOS"] != "3" || got["ITSEC"] != "12" || got["COSC"] != "2586" {
		t.Fatalf("%+v", page.Items)
	}
}

func TestJSMDesksSiteGarda(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jsm", "desks", "--site", "garda"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	if !strings.Contains(out.String(), "EOS") {
		t.Fatal(out.String())
	}
}

func TestJSMCommentPublicTrue(t *testing.T) {
	d, out, errw := testDeps()
	mem := d.JSM.(atlassian.JSMAPI).Memory
	code := Run([]string{"atlas", "jsm", "comment", "EOS-1", "--body", "customer note"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err, out.String())
	}
	if got["public"] != true {
		t.Fatalf("%v", got)
	}
	payload := mem.LastComment()
	if payload.Body != "customer note" || !payload.Public {
		t.Fatalf("%+v", payload)
	}
	req, err := mem.GetRequest(context.Background(), domain.GardaHostname, "EOS-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(req.Comments) != 1 || !req.Comments[0].Public || req.Comments[0].Body != "customer note" {
		t.Fatalf("%+v", req.Comments)
	}
}

func TestJSMCreateNoRaiseOnBehalfOf(t *testing.T) {
	d, out, errw := testDeps()
	mem := d.JSM.(atlassian.JSMAPI).Memory
	code := Run([]string{"atlas", "jsm", "create", "--desk", "3", "--type", "40", "--summary", "new request"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var created domain.CustomerRequest
	if err := json.Unmarshal(out.Bytes(), &created); err != nil {
		t.Fatal(err, out.String())
	}
	if created.Key != "EOS-2" {
		t.Fatalf("%+v", created)
	}
	if created.PortalURL != domain.PortalURL(domain.GardaHostname, "3", "EOS-2") {
		t.Fatalf("portal %q", created.PortalURL)
	}
	payload := mem.LastCreatePayload()
	if payload == nil {
		t.Fatal("missing create payload")
	}
	if _, ok := payload["raiseOnBehalfOf"]; ok {
		t.Fatalf("customers cannot raiseOnBehalfOf: %v", payload)
	}
	if payload["serviceDeskId"] != "3" || payload["requestTypeId"] != "40" {
		t.Fatalf("%v", payload)
	}
}

func TestJSMSiteSesamiIOIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "jsm", "desks", "--site", "sesami-io"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	s := errw.String()
	if !strings.Contains(s, "portal/1") && !strings.Contains(s, "atlas jira") {
		t.Fatal(s)
	}
	var e map[string]any
	if err := json.Unmarshal(errw.Bytes(), &e); err != nil {
		t.Fatal(s)
	}
	if e["class"] != "usage" {
		t.Fatalf("%v", e)
	}
}

func TestJSMSiteSesamidevelIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "jsm", "list", "--site", "sesamidevel"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
}

func TestJiraGetAndSearchSiteGardaRemainUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "jira", "get", "--site", "garda", "EOS-1"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	if !strings.Contains(errw.String(), "jsm") {
		t.Fatal(errw.String())
	}
	errw.Reset()
	code = Run([]string{"atlas", "jira", "search", "--site", "garda", "--jql", "project = EOS"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	if !strings.Contains(errw.String(), "atlas jsm") {
		t.Fatal(errw.String())
	}
}

func TestAuthStatusGardaRoleNoTokens(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "auth", "status"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	text := out.String()
	low := strings.ToLower(text)
	if strings.Contains(low, "token") || strings.Contains(low, "bearer") {
		t.Fatal(text)
	}
	var st map[string]any
	if err := json.Unmarshal(out.Bytes(), &st); err != nil {
		t.Fatal(err, text)
	}
	sites, _ := st["sites"].([]any)
	found := false
	for _, raw := range sites {
		m, _ := raw.(map[string]any)
		if m["alias"] != "garda" {
			continue
		}
		found = true
		if m["role"] != "jsm_customer" {
			t.Fatalf("%v", m)
		}
		if m["hostname"] != domain.GardaHostname {
			t.Fatalf("%v", m)
		}
		if _, ok := m["token"]; ok {
			t.Fatal(m)
		}
	}
	if !found {
		t.Fatalf("garda missing: %v", sites)
	}
}

func TestJSMGetSeedPortalURL(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jsm", "get", "EOS-1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var req domain.CustomerRequest
	if err := json.Unmarshal(out.Bytes(), &req); err != nil {
		t.Fatal(err, out.String())
	}
	if req.Key != "EOS-1" || req.DeskID != "3" {
		t.Fatalf("%+v", req)
	}
	want := "https://gardaworld.atlassian.net/servicedesk/customer/portal/3/EOS-1"
	if req.PortalURL != want {
		t.Fatalf("portal %q", req.PortalURL)
	}
}

func TestJSMUnknownVerbIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "jsm", "attach"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
}
