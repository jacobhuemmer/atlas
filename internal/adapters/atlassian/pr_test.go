package atlassian

import (
	"context"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMemoryPRGetSeed(t *testing.T) {
	mem := Seed()
	api := PRAPI{Memory: mem}
	pr, err := api.Get(context.Background(), domain.DefaultWorkspace(), "atlas", 1)
	if err != nil {
		t.Fatal(err)
	}
	if pr.Title == "" || pr.Workspace != domain.DefaultWorkspace() {
		t.Fatalf("%+v", pr)
	}
	if pr.State != "OPEN" {
		t.Fatalf("%+v", pr)
	}
	_, err = api.Get(context.Background(), domain.DefaultWorkspace(), "atlas", 99)
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
}
