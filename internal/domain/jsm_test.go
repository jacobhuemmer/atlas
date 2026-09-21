package domain

import "testing"

func TestPortalURL(t *testing.T) {
	got := PortalURL(GardaHostname, "3", "EOS-1")
	want := "https://gardaworld.atlassian.net/servicedesk/customer/portal/3/EOS-1"
	if got != want {
		t.Fatalf("%q != %q", got, want)
	}
}

func TestNormalizeJSMStatus(t *testing.T) {
	s, err := NormalizeJSMStatus("")
	if err != nil || s != JSMStatusAll {
		t.Fatalf("%q %v", s, err)
	}
	s, err = NormalizeJSMStatus("OPEN")
	if err != nil || s != JSMStatusOpen {
		t.Fatalf("%q %v", s, err)
	}
	if _, err := NormalizeJSMStatus("pending"); ExitOf(err) != ExitUsage {
		t.Fatal(err)
	}
}

func TestJSMRequestStatusQuery(t *testing.T) {
	if JSMRequestStatusQuery(JSMStatusOpen) != "OPEN_REQUESTS" {
		t.Fatal(JSMRequestStatusQuery(JSMStatusOpen))
	}
	if JSMRequestStatusQuery(JSMStatusClosed) != "CLOSED_REQUESTS" {
		t.Fatal(JSMRequestStatusQuery(JSMStatusClosed))
	}
	if JSMRequestStatusQuery(JSMStatusAll) != "ALL_REQUESTS" {
		t.Fatal(JSMRequestStatusQuery(JSMStatusAll))
	}
}
