package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMarkdownDescriptionBuildsRichADF(t *testing.T) {
	input := "## Objective\n\nStand up **uat-001** on `cl-iad-lznp-oke`.\n\n* First item\n* Second item\n\n1. Align IaC\n2. Deploy SES\n\n- [ ] Verify [runbook](https://example.test/runbook)\n\n```sh\natlas jira get SDO-588\n```"
	doc := adfFromMarkdown(input)
	blocks := asList(doc["content"])
	want := []string{"heading", "paragraph", "bulletList", "orderedList", "bulletList", "codeBlock"}
	if len(blocks) != len(want) {
		t.Fatalf("block count %d, want %d: %#v", len(blocks), len(want), doc)
	}
	for i, typ := range want {
		if str(asMap(blocks[i]), "type") != typ {
			t.Fatalf("block %d: %#v", i, blocks[i])
		}
	}
	paragraph := asList(asMap(blocks[1])["content"])
	var strong, code bool
	for _, raw := range paragraph {
		for _, mark := range asList(asMap(raw)["marks"]) {
			switch str(asMap(mark), "type") {
			case "strong":
				strong = true
			case "code":
				code = true
			}
		}
	}
	if !strong || !code {
		t.Fatalf("missing inline formatting: %#v", paragraph)
	}
	encoded, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	readback := markdownFromADF(decoded)
	for _, part := range []string{"## Objective", "**uat-001**", "`cl-iad-lznp-oke`", "1. Align IaC", "- [ ] Verify [runbook](https://example.test/runbook)", "```sh"} {
		if !strings.Contains(readback, part) {
			t.Fatalf("missing %q in %s", part, readback)
		}
	}
}

func TestJiraDescriptionMarkdownSentAsADFOnCreateAndEdit(t *testing.T) {
	var descriptions []map[string]any
	c := noSocketClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			desc := asMap(asMap(payload["fields"])["description"])
			descriptions = append(descriptions, desc)
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"key":"ABC-2"}`))
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"key": "ABC-2", "fields": map[string]any{
				"summary": "Example", "project": map[string]any{"key": "ABC"}, "description": descriptions[len(descriptions)-1],
			}})
		default:
			t.Fatal(r.Method)
		}
	}))
	j := Jira{Client: c}
	created, err := j.Create(context.Background(), "dev.example.atlassian.net", domain.CreateIssue{
		Project: "ABC", IssueType: "Task", Summary: "Example", Description: "## Objective\n\n**Bold**",
	}, false)
	if err != nil || !strings.Contains(created.Description, "## Objective") {
		t.Fatalf("%+v %v", created, err)
	}
	edited, err := j.Edit(context.Background(), "dev.example.atlassian.net", "ABC-2", map[string]any{"description": "## Updated\n\n- item"}, false)
	if err != nil || !strings.Contains(edited.Description, "## Updated") {
		t.Fatalf("%+v %v", edited, err)
	}
	if len(descriptions) != 2 || str(asMap(asList(descriptions[0]["content"])[0]), "type") != "heading" || str(asMap(asList(descriptions[1]["content"])[0]), "type") != "heading" {
		t.Fatalf("descriptions %#v", descriptions)
	}
}
