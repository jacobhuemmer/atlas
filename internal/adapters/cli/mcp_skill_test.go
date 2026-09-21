package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPSkillResource(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	ctx := context.Background()

	list, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, resource := range list.Resources {
		if resource.URI == skillURI {
			found = true
			if resource.MIMEType != "text/markdown" {
				t.Fatalf("skill MIME type %q", resource.MIMEType)
			}
		}
	}
	if !found {
		t.Fatalf("missing %s in %#v", skillURI, list.Resources)
	}

	read, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: skillURI})
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Contents) != 1 {
		t.Fatalf("skill contents count %d", len(read.Contents))
	}
	content := read.Contents[0]
	if content.URI != skillURI || content.MIMEType != "text/markdown" {
		t.Fatalf("skill content metadata %#v", content)
	}

	help, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "atlas_help", Arguments: helpIn{Topic: "atlas"},
	})
	if err != nil || help.IsError {
		t.Fatal(err, toolText(t, help))
	}
	resourceBody := strings.TrimSpace(content.Text)
	helpBody := strings.TrimSpace(toolText(t, help))
	wantBody := strings.TrimSpace(skillMarkdown)
	if resourceBody != wantBody || helpBody != wantBody {
		t.Fatal("skill resource and help topic differ from embedded markdown")
	}
	for _, want := range []string{"atlas_run", "write_opt_in"} {
		if !strings.Contains(resourceBody, want) {
			t.Fatalf("skill body missing %q", want)
		}
	}

	prompts, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompts.Prompts) != 4 {
		t.Fatalf("prompt count %d", len(prompts.Prompts))
	}
	for _, prompt := range prompts.Prompts {
		if prompt.Name == "atlas" {
			t.Fatal("embedded skill must not be an MCP prompt")
		}
	}

	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 3 {
		t.Fatalf("tool count %d", len(tools.Tools))
	}
}

func TestMCPUnknownResourceDoesNotReturnSkill(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	read, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "atlas://nope"})
	if err == nil {
		t.Fatal("unknown resource read succeeded")
	}
	if read != nil {
		for _, content := range read.Contents {
			if strings.Contains(content.Text, "write_opt_in") {
				t.Fatal("unknown resource returned the skill body")
			}
		}
	}
}

func TestReadSkillRejectsForeignURI(t *testing.T) {
	result, err := readSkill(context.Background(), &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{URI: "atlas://nope"},
	})
	if err == nil {
		t.Fatal("foreign URI read succeeded")
	}
	if result != nil {
		t.Fatal("foreign URI returned resource contents")
	}
}
