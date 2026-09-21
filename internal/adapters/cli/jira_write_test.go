package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/adapters/atlassian"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestJiraCreatePersistsAndGetFindsKey(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jira", "create", "--project", "SDO", "--type", "Task", "--summary", "Phase 3 write"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var created domain.Issue
	if err := json.Unmarshal(out.Bytes(), &created); err != nil {
		t.Fatal(err, out.String())
	}
	if created.Key != "SDO-2" {
		t.Fatalf("key %q", created.Key)
	}
	if created.Site != "sesamidevel.atlassian.net" {
		t.Fatalf("site %q", created.Site)
	}
	if created.Assignee != domain.DefaultAssigneeAccountID {
		t.Fatalf("assignee %q", created.Assignee)
	}
	out.Reset()
	errw.Reset()
	code = Run([]string{"atlas", "jira", "get", created.Key}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var got domain.Issue
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err, out.String())
	}
	if got.Summary != "Phase 3 write" || got.Key != created.Key {
		t.Fatalf("%+v", got)
	}
}

func TestJiraCommentDryRunDoesNotAppend(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jira", "comment", "SDO-1", "--body", "do not persist", "--dry-run"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var preview map[string]any
	if err := json.Unmarshal(out.Bytes(), &preview); err != nil {
		t.Fatal(err, out.String())
	}
	if preview["dry_run"] != true {
		t.Fatalf("%v", preview)
	}
	if preview["namespace"] != "jira" || preview["verb"] != "comment" {
		t.Fatalf("%v", preview)
	}
	out.Reset()
	errw.Reset()
	code = Run([]string{"atlas", "jira", "get", "SDO-1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var iss domain.Issue
	if err := json.Unmarshal(out.Bytes(), &iss); err != nil {
		t.Fatal(err, out.String())
	}
	if len(iss.Comments) != 0 {
		t.Fatalf("comments %v", iss.Comments)
	}
}

func TestJiraCommentPersists(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jira", "comment", "SDO-1", "--body", "ship it"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	out.Reset()
	errw.Reset()
	code = Run([]string{"atlas", "jira", "get", "SDO-1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var iss domain.Issue
	if err := json.Unmarshal(out.Bytes(), &iss); err != nil {
		t.Fatal(err, out.String())
	}
	if len(iss.Comments) != 1 || iss.Comments[0] != "ship it" {
		t.Fatalf("comments %v", iss.Comments)
	}
}

func TestJiraTransitionSDO1Done(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jira", "transition", "SDO-1", "--name", "Done"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var iss domain.Issue
	if err := json.Unmarshal(out.Bytes(), &iss); err != nil {
		t.Fatal(err, out.String())
	}
	if iss.Status != "Done" {
		t.Fatalf("status %q", iss.Status)
	}
	out.Reset()
	errw.Reset()
	code = Run([]string{"atlas", "jira", "get", "SDO-1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	if err := json.Unmarshal(out.Bytes(), &iss); err != nil {
		t.Fatal(err, out.String())
	}
	if iss.Status != "Done" {
		t.Fatalf("status %q", iss.Status)
	}
}

func TestJiraTransitionUnknownNameIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "jira", "transition", "SDO-1", "--name", "Yeet"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	var e map[string]any
	if err := json.Unmarshal(errw.Bytes(), &e); err != nil {
		t.Fatal(errw.String())
	}
	if e["class"] != "usage" {
		t.Fatalf("%v", e)
	}
}

func TestJiraCreateSDPStoryIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	mem := d.Jira.(atlassian.JiraAPI).Memory
	before := mem.IssueCount()
	code := Run([]string{"atlas", "jira", "create", "--project", "SDP", "--type", "Story", "--summary", "nope"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	if mem.IssueCount() != before {
		t.Fatal("seed mutated")
	}
	s := errw.String()
	if !strings.Contains(s, "Story") {
		t.Fatal(s)
	}
}

func TestJiraEditFields(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jira", "edit", "SDO-1", "--fields", `{"summary":"renamed"}`}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var iss domain.Issue
	if err := json.Unmarshal(out.Bytes(), &iss); err != nil {
		t.Fatal(err, out.String())
	}
	if iss.Summary != "renamed" {
		t.Fatalf("%+v", iss)
	}
}
