package domain

import "testing"

func TestLookupSites(t *testing.T) {
	if len(Sites) != 3 {
		t.Fatalf("count %d", len(Sites))
	}
	s, err := Lookup("sesamidevel")
	if err != nil || s.Hostname != "sesamidevel.atlassian.net" || s.Role != "licensed" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = Lookup("8986e79a-6c2a-4e0d-bb65-4a57e0f68db9")
	if err != nil || s.Alias != "sesami-io" {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = Lookup("gardaworld.atlassian.net")
	if err != nil || s.Role != "jsm_customer" {
		t.Fatalf("%+v %v", s, err)
	}
	if _, err := Lookup("nope"); ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
}

func TestSignedOutHasThreeUnusableSites(t *testing.T) {
	s := SignedOut()
	if s.SignedIn || s.SessionUsable {
		t.Fatalf("%+v", s)
	}
	if len(s.Sites) != 3 {
		t.Fatalf("%+v", s.Sites)
	}
	for _, st := range s.Sites {
		if st.Usable {
			t.Fatal(st)
		}
	}
}
