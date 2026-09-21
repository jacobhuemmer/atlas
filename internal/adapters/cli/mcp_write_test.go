package cli

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/atlas/internal/adapters/atlassian"
	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMCPJiraCreateWriteGate(t *testing.T) {
	d, _, _ := testDeps()
	mem := d.Jira.(atlassian.JiraAPI).Memory
	before := mem.IssueCount()
	cs := connectMCP(t, d)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jira", Verb: "create",
			Flags: map[string]any{"project": "SDO", "type": "Task", "summary": "gated"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	var preview map[string]any
	if err := json.Unmarshal([]byte(toolText(t, res)), &preview); err != nil {
		t.Fatal(toolText(t, res))
	}
	if preview["dry_run"] != true {
		t.Fatal(toolText(t, res))
	}
	if preview["namespace"] != "jira" || preview["verb"] != "create" {
		t.Fatal(toolText(t, res))
	}
	if preview["project"] != "SDO" || preview["summary"] != "gated" {
		t.Fatal(toolText(t, res))
	}
	if mem.IssueCount() != before {
		t.Fatal("seed issue count changed on dry-run")
	}

	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jira", Verb: "create", WriteOptIn: true,
			Flags: map[string]any{"project": "SDO", "type": "Task", "summary": "gated"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	var created domain.Issue
	if err := json.Unmarshal([]byte(toolText(t, res)), &created); err != nil {
		t.Fatal(toolText(t, res))
	}
	if created.Key == "" || created.Summary != "gated" {
		t.Fatalf("%+v", created)
	}
	if mem.IssueCount() != before+1 {
		t.Fatalf("count %d", mem.IssueCount())
	}

	got, err := mem.Get(context.Background(), created.Site, created.Key)
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != created.Key || got.Summary != "gated" {
		t.Fatalf("%+v", got)
	}
}
