package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestPRGetSeed(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "pr", "get", "--repo", "atlas", "--id", "1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var pr domain.PullRequest
	if err := json.Unmarshal(out.Bytes(), &pr); err != nil {
		t.Fatal(err, out.String())
	}
	if pr.Title == "" {
		t.Fatalf("%+v", pr)
	}
	if pr.Workspace != domain.DefaultWorkspace {
		t.Fatalf("workspace %q", pr.Workspace)
	}
	if pr.Repo != "atlas" || pr.ID != 1 {
		t.Fatalf("%+v", pr)
	}
	if pr.State != "OPEN" {
		t.Fatalf("state %q", pr.State)
	}
}

func TestPRListSeed(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "pr", "list", "--repo", "atlas"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var page domain.PullRequestList
	if err := json.Unmarshal(out.Bytes(), &page); err != nil {
		t.Fatal(err, out.String())
	}
	if page.Workspace != domain.DefaultWorkspace || page.Repo != "atlas" {
		t.Fatalf("%+v", page)
	}
	if page.Count == 0 {
		t.Fatal(page)
	}
	found := false
	for _, it := range page.Items {
		if it.ID == 1 && it.Workspace == domain.DefaultWorkspace {
			found = true
		}
	}
	if !found {
		t.Fatalf("%+v", page.Items)
	}
}

func TestPRDiffSeed(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "pr", "diff", "--repo", "atlas", "--id", "1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var diff domain.PullRequestDiff
	if err := json.Unmarshal(out.Bytes(), &diff); err != nil {
		t.Fatal(err, out.String())
	}
	if diff.ID != 1 || diff.Workspace != domain.DefaultWorkspace {
		t.Fatalf("%+v", diff)
	}
	if !strings.Contains(diff.Diff, "diff --git") {
		t.Fatalf("%q", diff.Diff)
	}
}

func TestPRGetUnknownIsNotFound(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "pr", "get", "--repo", "atlas", "--id", "999"}, d)
	if code != domain.ExitNotFound {
		t.Fatal(code, errw.String())
	}
}

func TestPRDeleteIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "pr", "delete", "--repo", "atlas", "--id", "1"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
	s := errw.String()
	if !strings.Contains(s, "unknown pr verb") {
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
