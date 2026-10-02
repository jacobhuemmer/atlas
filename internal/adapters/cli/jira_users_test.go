package cli

import (
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestJiraUsersFindsAccountIDOnIssueSite(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jira", "users", "--query", "alex", "--issue", "SDO-1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var result domain.UserSearchResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err, out.String())
	}
	if result.Site != "sesamidevel.atlassian.net" || result.Issue != "SDO-1" || result.Count != 1 || result.Users[0].AccountID != "acct-alex" {
		t.Fatalf("%+v", result)
	}
	out.Reset()
	code = Run([]string{"atlas", "jira", "edit", "SDO-1", "--assignee", result.Users[0].AccountID}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var assigned domain.Issue
	if err := json.Unmarshal(out.Bytes(), &assigned); err != nil {
		t.Fatal(err, out.String())
	}
	if assigned.Assignee != "acct-alex" {
		t.Fatalf("assignee %q", assigned.Assignee)
	}
}

func TestJiraUsersRequiresQueryAndOneScope(t *testing.T) {
	d, _, errw := testDeps()
	for _, args := range [][]string{
		{"atlas", "jira", "users", "--project", "SDO"},
		{"atlas", "jira", "users", "--query", "alex", "--project", "SDO", "--issue", "SDO-1"},
	} {
		errw.Reset()
		if code := Run(args, d); code != domain.ExitUsage {
			t.Fatalf("%v: %d %s", args, code, errw.String())
		}
	}
}
