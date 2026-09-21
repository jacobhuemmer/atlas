package cli

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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
		Name: "atlas_run", Arguments: runIn{Namespace: "jira", Verb: "get", Args: []string{"SDO-1"}},
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
