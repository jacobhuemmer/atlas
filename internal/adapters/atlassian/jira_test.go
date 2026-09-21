package atlassian

import (
	"context"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMemoryGetAndSearchIsolation(t *testing.T) {
	mem := Seed()
	api := JiraAPI{Memory: mem}
	iss, err := api.Get(context.Background(), host("sesamidevel"), "SDO-1")
	if err != nil || iss.Key != "SDO-1" || iss.Site != host("sesamidevel") {
		t.Fatalf("%+v %v", iss, err)
	}
	_, err = api.Get(context.Background(), host("sesami-io"), "SDO-1")
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
	page, err := api.Search(context.Background(), host("sesami-io"), "project = CAB")
	if err != nil {
		t.Fatal(err)
	}
	if page.Site != host("sesami-io") || page.Count == 0 {
		t.Fatalf("%+v", page)
	}
	for _, it := range page.Items {
		if it.Site != host("sesami-io") {
			t.Fatal(it)
		}
	}
}
