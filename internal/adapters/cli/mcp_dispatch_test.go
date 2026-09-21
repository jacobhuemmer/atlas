package cli

import (
	"reflect"
	"testing"
)

func TestFlagMapToArgs(t *testing.T) {
	got, err := FlagMapToArgs("jira", "search", nil, map[string]any{"jql": "project = CAB", "site": "sesami-io"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"atlas", "jira", "search", "--jql", "project = CAB", "--site", "sesami-io"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v != %v", got, want)
	}
}

func TestWriteGateInjectsDryRun(t *testing.T) {
	writes := [][2]string{
		{"jira", "create"},
		{"jira", "edit"},
		{"jira", "comment"},
		{"jira", "transition"},
		{"jira", "link"},
		{"confluence", "create"},
		{"confluence", "update"},
		{"pr", "create"},
		{"pr", "comment"},
		{"pr", "merge"},
	}
	for _, w := range writes {
		if !isWrite(w[0], w[1]) {
			t.Fatalf("isWrite %s %s", w[0], w[1])
		}
		args, err := buildRunArgs(w[0], w[1], []string{"SDO-1"}, map[string]any{"dry-run": false}, false)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, a := range args {
			if a == "--dry-run" {
				found = true
			}
		}
		if !found {
			t.Fatal(args)
		}
		args, err = buildRunArgs(w[0], w[1], nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		for _, a := range args {
			if a == "--dry-run" {
				t.Fatal(args)
			}
		}
	}
	if isWrite("jira", "get") || isWrite("jira", "search") || isWrite("confluence", "get") || isWrite("confluence", "search") || isWrite("confluence", "delete") || isWrite("pr", "get") || isWrite("pr", "list") || isWrite("pr", "diff") || isWrite("pr", "delete") {
		t.Fatal("reads and delete are not writes")
	}
	args, err := buildRunArgs("jira", "get", []string{"SDO-1"}, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range args {
		if a == "--dry-run" {
			t.Fatal(args)
		}
	}
}
