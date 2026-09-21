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

func TestMCPJSMCreateWriteGate(t *testing.T) {
	d, _, _ := testDeps()
	mem := d.JSM.(atlassian.JSMAPI).Memory
	before := mem.RequestCount()
	cs := connectMCP(t, d)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jsm", Verb: "create",
			Flags: map[string]any{"desk": "3", "type": "40", "summary": "gated jsm"},
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
	if preview["namespace"] != "jsm" || preview["verb"] != "create" {
		t.Fatal(toolText(t, res))
	}
	if mem.RequestCount() != before {
		t.Fatal("seed request count changed on dry-run")
	}

	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jsm", Verb: "create", WriteOptIn: true,
			Flags: map[string]any{"desk": "3", "type": "40", "summary": "gated jsm"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	var created domain.CustomerRequest
	if err := json.Unmarshal([]byte(toolText(t, res)), &created); err != nil {
		t.Fatal(toolText(t, res))
	}
	if created.Key == "" || created.Summary != "gated jsm" {
		t.Fatalf("%+v", created)
	}
	if mem.RequestCount() != before+1 {
		t.Fatalf("count %d", mem.RequestCount())
	}
}

func TestMCPJSMCommentWriteGate(t *testing.T) {
	d, _, _ := testDeps()
	mem := d.JSM.(atlassian.JSMAPI).Memory
	cs := connectMCP(t, d)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jsm", Verb: "comment",
			Args:  []string{"EOS-1"},
			Flags: map[string]any{"body": "gated comment"},
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
	got, err := mem.GetRequest(context.Background(), domain.GardaHostname, "EOS-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Comments) != 0 {
		t.Fatalf("comment persisted without write_opt_in %+v", got.Comments)
	}

	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jsm", Verb: "comment", WriteOptIn: true,
			Args:  []string{"EOS-1"},
			Flags: map[string]any{"body": "gated comment"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	got, err = mem.GetRequest(context.Background(), domain.GardaHostname, "EOS-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Comments) != 1 || !got.Comments[0].Public || got.Comments[0].Body != "gated comment" {
		t.Fatalf("%+v", got.Comments)
	}
}

func TestMCPJSMTransitionWriteGate(t *testing.T) {
	d, _, _ := testDeps()
	mem := d.JSM.(atlassian.JSMAPI).Memory
	cs := connectMCP(t, d)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jsm", Verb: "transition",
			Args:  []string{"EOS-1"},
			Flags: map[string]any{"id": "21"},
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
	got, err := mem.GetRequest(context.Background(), domain.GardaHostname, "EOS-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.StatusCategory != domain.JSMCategoryOpen {
		t.Fatalf("transition persisted without write_opt_in %+v", got)
	}

	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jsm", Verb: "transition", WriteOptIn: true,
			Args:  []string{"EOS-1"},
			Flags: map[string]any{"id": "21"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	got, err = mem.GetRequest(context.Background(), domain.GardaHostname, "EOS-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.StatusCategory != domain.JSMCategoryClosed {
		t.Fatalf("%+v", got)
	}
}

func TestMCPJiraLinkWriteGate(t *testing.T) {
	d, _, _ := testDeps()
	mem := d.Jira.(atlassian.JiraAPI).Memory
	cs := connectMCP(t, d)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jira", Verb: "link",
			Args: []string{"SDO-1", "SDP-2"},
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
	if preview["namespace"] != "jira" || preview["verb"] != "link" {
		t.Fatal(toolText(t, res))
	}
	if preview["type"] != domain.DefaultLinkType {
		t.Fatal(toolText(t, res))
	}
	got, err := mem.Get(context.Background(), "sesamidevel.atlassian.net", "SDO-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Links) != 0 {
		t.Fatalf("link persisted without write_opt_in %+v", got.Links)
	}

	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "jira", Verb: "link", WriteOptIn: true,
			Args: []string{"SDO-1", "SDP-2"},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	got, err = mem.Get(context.Background(), "sesamidevel.atlassian.net", "SDO-1")
	if err != nil {
		t.Fatal(err)
	}
	if !hasRelatesLink(got, "SDO-1", "SDP-2") {
		t.Fatalf("%+v", got.Links)
	}
}
