package atlassian

import (
	"context"
	"testing"

	"github.com/masonhuemmer/atlas/internal/domain"
)

func TestMemoryGetAndSearchIsolation(t *testing.T) {
	mem := Seed()
	api := JiraAPI{Memory: mem}
	iss, err := api.Get(context.Background(), hostDevel, "SDO-1")
	if err != nil || iss.Key != "SDO-1" || iss.Site != hostDevel {
		t.Fatalf("%+v %v", iss, err)
	}
	_, err = api.Get(context.Background(), hostIO, "SDO-1")
	if domain.ExitOf(err) != domain.ExitNotFound {
		t.Fatal(err)
	}
	page, err := api.Search(context.Background(), hostIO, "project = CAB")
	if err != nil {
		t.Fatal(err)
	}
	if page.Site != hostIO || page.Count == 0 {
		t.Fatalf("%+v", page)
	}
	for _, it := range page.Items {
		if it.Site != hostIO {
			t.Fatal(it)
		}
	}
}
