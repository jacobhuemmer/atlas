package atlassian

import (
	"context"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMemoryConfluenceSearchIsolation(t *testing.T) {
	mem := Seed()
	api := ConfluenceAPI{Memory: mem}
	page, err := api.Search(context.Background(), host("sesami-io"), "space = CCAB AND type = page")
	if err != nil {
		t.Fatal(err)
	}
	if page.Site != host("sesami-io") || page.Count == 0 {
		t.Fatalf("%+v", page)
	}
	found := false
	for _, it := range page.Items {
		if it.Site != host("sesami-io") {
			t.Fatal(it)
		}
		if it.Title == "CAB-109" {
			found = true
		}
	}
	if !found {
		t.Fatal(page.Items)
	}
	got, err := api.Get(context.Background(), host("sesami-io"), seedPageID)
	if err != nil || got.Title != "CAB-109" {
		t.Fatalf("%+v %v", got, err)
	}
	_, err = api.Get(context.Background(), host("sesamidevel"), seedPageID)
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
}
