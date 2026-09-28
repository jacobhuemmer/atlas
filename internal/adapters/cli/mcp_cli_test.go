package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMCPHelpAndHuman(t *testing.T) {
	d, out, errw := testDeps()
	if c := Run([]string{"atlas", "--help"}, d); c != 0 {
		t.Fatal(c)
	}
	if !strings.Contains(out.String(), "mcp") {
		t.Fatal(out.String())
	}
	out.Reset()
	if c := Run([]string{"atlas", "mcp", "--help"}, d); c != 0 {
		t.Fatal(errw.String())
	}
	s := out.String()
	for _, want := range []string{"atlas_status", "atlas_help", "atlas_read", "atlas_write", "write_opt_in", "serve", "jira-search", "confluence-write", "atlas://skill", "topic=atlas"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
	out.Reset()
	if c := Run([]string{"atlas", "mcp", "serve", "--help"}, d); c != 0 {
		t.Fatal(c, errw.String())
	}
	if c := Run([]string{"atlas", "mcp", "nope"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
	if c := Run([]string{"atlas", "--human", "mcp", "serve"}, d); c != domain.ExitUsage {
		t.Fatal(c, errw.String())
	}
}

func TestMCPRunLoginAndMCPNamespaceRefused(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	for _, tool := range []string{"atlas_read", "atlas_write"} {
		for _, in := range []readIn{{Namespace: "auth", Verb: "login"}, {Namespace: "auth", Verb: "logout"}, {Namespace: "mcp", Verb: "serve"}} {
			res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: in})
			if err != nil {
				t.Fatal(err)
			}
			assertClass(t, res, "usage")
		}
	}
}

func assertClass(t *testing.T, res *mcp.CallToolResult, class string) {
	t.Helper()
	if !res.IsError {
		t.Fatal(toolText(t, res))
	}
	var e map[string]any
	if err := json.Unmarshal([]byte(toolText(t, res)), &e); err != nil {
		t.Fatal(toolText(t, res))
	}
	if e["class"] != class {
		t.Fatalf("%v", e)
	}
}
