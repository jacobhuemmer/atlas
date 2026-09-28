package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func connectMCP(t *testing.T, d Deps) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := NewMCPServer(d).Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func toolText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if res == nil || len(res.Content) == 0 {
		t.Fatal("no content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("%T", res.Content[0])
	}
	return tc.Text
}

func TestToolsListCompactCatalog(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	list, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tools) != 4 {
		t.Fatalf("count %d", len(list.Tools))
	}
	got := map[string]string{}
	for _, tl := range list.Tools {
		got[tl.Name] = tl.Description
	}
	for _, name := range []string{"atlas_status", "atlas_help", "atlas_read", "atlas_write"} {
		if got[name] == "" {
			t.Fatalf("missing %s in %#v", name, got)
		}
	}
	if !strings.Contains(got["atlas_read"], "jira-search") {
		t.Fatal(got["atlas_read"])
	}
	if !strings.Contains(got["atlas_write"], "write_opt_in") {
		t.Fatal(got["atlas_write"])
	}
}

// Clients such as Codex run read-only tools without a prompt, so the hints
// must be true only for tools that cannot change a workload.
func TestToolAnnotations(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	list, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tl := range list.Tools {
		a := tl.Annotations
		if a == nil {
			t.Fatalf("%s has no annotations", tl.Name)
		}
		switch tl.Name {
		case "atlas_status", "atlas_help", "atlas_read":
			if !a.ReadOnlyHint {
				t.Fatalf("%s not read-only", tl.Name)
			}
		case "atlas_write":
			if a.ReadOnlyHint || a.DestructiveHint == nil || !*a.DestructiveHint {
				t.Fatalf("atlas_write hints %+v", a)
			}
		default:
			t.Fatalf("unexpected tool %s", tl.Name)
		}
	}
}

func TestAtlasRunRemoved(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: readIn{Namespace: "jira", Verb: "get", Args: []string{"SDO-1"}},
	})
	if err == nil && (res == nil || !res.IsError) {
		t.Fatal("atlas_run still answers")
	}
}
