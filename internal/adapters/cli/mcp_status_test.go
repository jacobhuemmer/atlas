package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/masonhuemmer/atlas/internal/app/auth"
)

func TestMCPStatusSignedOutAndIn(t *testing.T) {
	d, _, _ := testDeps()
	cs := connectMCP(t, d)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "atlas_status"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatal(toolText(t, res))
	}
	text := toolText(t, res)
	low := strings.ToLower(text)
	if strings.Contains(low, "token") || strings.Contains(low, "bearer") {
		t.Fatal(text)
	}
	var st map[string]any
	if err := json.Unmarshal([]byte(text), &st); err != nil {
		t.Fatal(err, text)
	}
	if st["signed_in"] != false || st["session_usable"] != false {
		t.Fatal(st)
	}
	sites, _ := st["sites"].([]any)
	if len(sites) != 3 {
		t.Fatalf("sites %v", sites)
	}
	for _, raw := range sites {
		m, _ := raw.(map[string]any)
		if m["usable"] != false {
			t.Fatal(m)
		}
		if _, ok := m["token"]; ok {
			t.Fatal(m)
		}
	}

	if err := auth.PutSite(d.Store, "sesamidevel", auth.Cred{Email: "a@b.c", Token: "secret-token"}); err != nil {
		t.Fatal(err)
	}
	cs = connectMCP(t, d)
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "atlas_status"})
	if err != nil || res.IsError {
		t.Fatal(err, toolText(t, res))
	}
	text = toolText(t, res)
	if strings.Contains(strings.ToLower(text), "token") || strings.Contains(text, "secret-token") {
		t.Fatal(text)
	}
	if err := json.Unmarshal([]byte(text), &st); err != nil {
		t.Fatal(text)
	}
	if st["signed_in"] != true || st["session_usable"] != true {
		t.Fatal(st)
	}
}
