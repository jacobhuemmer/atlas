package cli

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/masonhuemmer/atlas/internal/adapters/atlassian"
	"github.com/masonhuemmer/atlas/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPRCreateMinimumTitleSource(t *testing.T) {
	d, out, errw := testDeps()
	code := Run([]string{"atlas", "pr", "create", "--repo", "atlas", "--title", "Phase 6 create", "--source", "feature/pr-cli"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var created domain.PullRequest
	if err := json.Unmarshal(out.Bytes(), &created); err != nil {
		t.Fatal(err, out.String())
	}
	if created.ID != 2 {
		t.Fatalf("id %d", created.ID)
	}
	if created.Title != "Phase 6 create" || created.Source != "feature/pr-cli" {
		t.Fatalf("%+v", created)
	}
	if created.Target != domain.DefaultTargetBranch {
		t.Fatalf("target %q", created.Target)
	}
	if created.Workspace != domain.DefaultWorkspace || created.State != "OPEN" {
		t.Fatalf("%+v", created)
	}
	out.Reset()
	errw.Reset()
	code = Run([]string{"atlas", "pr", "get", "--repo", "atlas", "--id", "2"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
}

func TestPRCreateMissingSourceIsUsage(t *testing.T) {
	d, _, errw := testDeps()
	code := Run([]string{"atlas", "pr", "create", "--repo", "atlas", "--title", "no source"}, d)
	if code != domain.ExitUsage {
		t.Fatal(code, errw.String())
	}
}

func TestPRCreateDryRunDoesNotPersist(t *testing.T) {
	d, out, errw := testDeps()
	mem := d.PR.(atlassian.PRAPI).Memory
	before := mem.PRCount()
	code := Run([]string{"atlas", "pr", "create", "--repo", "atlas", "--title", "dry", "--source", "feature/x", "--dry-run"}, d)
	if code != domain.ExitOK {
		t.Fatal(code, errw.String())
	}
	var preview map[string]any
	if err := json.Unmarshal(out.Bytes(), &preview); err != nil {
		t.Fatal(err, out.String())
	}
	if preview["dry_run"] != true || preview["namespace"] != "pr" || preview["verb"] != "create" {
		t.Fatalf("%v", preview)
	}
	if mem.PRCount() != before {
		t.Fatal("seed PR count changed on dry-run")
	}
}

func TestMCPPRMergeWriteGate(t *testing.T) {
	d, _, _ := testDeps()
	mem := d.PR.(atlassian.PRAPI).Memory
	cs := connectMCP(t, d)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "pr", Verb: "merge",
			Flags: map[string]any{"repo": "atlas", "id": float64(1)},
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
	if preview["namespace"] != "pr" || preview["verb"] != "merge" {
		t.Fatal(toolText(t, res))
	}
	got, err := mem.GetPR(context.Background(), domain.DefaultWorkspace, "atlas", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "OPEN" {
		t.Fatalf("seed PR merged without write_opt_in %+v", got)
	}

	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "atlas_run", Arguments: runIn{
			Namespace: "pr", Verb: "merge", WriteOptIn: true,
			Flags: map[string]any{"repo": "atlas", "id": float64(1)},
		},
	})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	var merged domain.PullRequest
	if err := json.Unmarshal([]byte(toolText(t, res)), &merged); err != nil {
		t.Fatal(toolText(t, res))
	}
	if merged.State != "MERGED" {
		t.Fatalf("%+v", merged)
	}
	got, err = mem.GetPR(context.Background(), domain.DefaultWorkspace, "atlas", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "MERGED" {
		t.Fatalf("%+v", got)
	}
}
