package cli

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/atlas/internal/adapters/atlassian"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMCPRunJiraGetParity(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "jira", "get", "SDO-1"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var cliIssue domain.Issue
	if err := json.Unmarshal(out.Bytes(), &cliIssue); err != nil {
		t.Fatal(err, out.String())
	}

	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_read", Arguments: readIn{Namespace: "jira", Verb: "get", Args: []string{"SDO-1"}},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	var mcpIssue domain.Issue
	if err := json.Unmarshal([]byte(toolText(t, res)), &mcpIssue); err != nil {
		t.Fatal(toolText(t, res))
	}
	if !reflect.DeepEqual(mcpIssue, cliIssue) {
		t.Fatalf("%+v vs %+v", mcpIssue, cliIssue)
	}
	if mcpIssue.Key != "SDO-1" || mcpIssue.Site != "sesamidevel.atlassian.net" {
		t.Fatalf("%+v", mcpIssue)
	}
	if mcpIssue.BrowseURL != "https://sesamidevel.atlassian.net/browse/SDO-1" {
		t.Fatalf("%+v", mcpIssue)
	}
}

func TestMCPReadRefusesWrites(t *testing.T) {
	d, _, _ := testDeps()
	mem := d.Jira.(atlassian.JiraAPI).Memory
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_read", Arguments: readIn{
			Namespace: "jira", Verb: "comment",
			Args:  []string{"SDO-1"},
			Flags: map[string]any{"body": "via read"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertClass(t, res, "usage")
	if !strings.Contains(toolText(t, res), "atlas_write") {
		t.Fatal(toolText(t, res))
	}
	got, err := mem.Get(context.Background(), "sesamidevel.atlassian.net", "SDO-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Comments) != 0 {
		t.Fatalf("comment persisted through atlas_read %+v", got.Comments)
	}
}

// atlas_read is an allowlist: a verb it does not know is refused, not run,
// so a new write verb can never ride in on the read-only tool.
func TestMCPReadRefusesUnknownVerbs(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	for _, in := range []readIn{
		{Namespace: "confluence", Verb: "delete", Args: []string{"1"}},
		{Namespace: "pr", Verb: "decline"},
		{Namespace: "nope", Verb: "get"},
	} {
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "atlas_read", Arguments: in})
		if err != nil {
			t.Fatal(err)
		}
		assertClass(t, res, "usage")
	}
}

func TestMCPWriteRefusesReads(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_write", Arguments: writeIn{Namespace: "jira", Verb: "get", Args: []string{"SDO-1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertClass(t, res, "usage")
	if !strings.Contains(toolText(t, res), "atlas_read") {
		t.Fatal(toolText(t, res))
	}
}
